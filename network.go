package run

import (
	"encoding/json"
	"net"
	"net/http"
)

type IPAddressNotFoundError struct{}

func (e IPAddressNotFoundError) Error() string {
	return "run: RFC1918 address not found"
}

type NetworkInterface struct {
	Name         string   `json:"name"`
	Index        int      `json:"index"`
	HardwareAddr string   `json:"hardware_address"`
	IPAddresses  []string `json:"ip_addresses"`
}

// IPAddress returns the RFC 1918 IP address assigned to the Cloud Run instance.
func IPAddress(interfaces ...net.Interface) (string, error) {
	ips, err := netIPAddresses(interfaces...)
	if err != nil {
		return "", err
	}

	blocks := rfc1918Blocks()
	for _, ip := range ips {
		// Skip 192.168.1.1 as this is address should not be exposed to Cloud Run instances.
		if ip.String() == "192.168.1.1" {
			continue
		}

		for _, block := range blocks {
			if block.Contains(ip) {
				return ip.String(), nil
			}
		}
	}
	return "", IPAddressNotFoundError{}
}

func IPAddresses() ([]string, error) {
	ipAddresses := make([]string, 0)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipAddresses = append(ipAddresses, ipnet.IP.String())
			}
		}
	}
	return ipAddresses, nil
}

// NetworkInterfaces returns a list of network interfaces attached to the
// Cloud Run instance.
func NetworkInterfaces(interfaces ...net.Interface) ([]NetworkInterface, error) {
	var err error
	interfaces, err = netInterfaces(interfaces...)
	if err != nil {
		return nil, err
	}

	networkInterfaces := make([]NetworkInterface, 0)
	for _, i := range interfaces {
		var err error
		netips, err := netIPAddresses(i)
		if err != nil {
			return networkInterfaces, err
		}

		ips := make([]string, 0)
		for _, netIP := range netips {
			ips = append(ips, netIP.String())
		}

		ni := NetworkInterface{
			Name:         i.Name,
			Index:        i.Index,
			HardwareAddr: i.HardwareAddr.String(),
			IPAddresses:  ips,
		}

		networkInterfaces = append(networkInterfaces, ni)
	}

	return networkInterfaces, nil
}

func netInterfaces(interfaces ...net.Interface) ([]net.Interface, error) {
	if len(interfaces) == 0 {
		return net.Interfaces()
	}

	return interfaces, nil
}

func netIPAddresses(interfaces ...net.Interface) ([]net.IP, error) {
	ips := make([]net.IP, 0)

	ifs, err := netInterfaces(interfaces...)
	if err != nil {
		return nil, err
	}

	for _, i := range ifs {
		var err error
		addrs, err := i.Addrs()
		if err != nil {
			return ips, err
		}
		for _, address := range addrs {
			ip, _, err := net.ParseCIDR(address.String())
			if err != nil {
				return ips, err
			}
			ips = append(ips, net.ParseIP(ip.String()))
		}
	}

	return ips, nil
}

func rfc1918Blocks() []*net.IPNet {
	blocks := make([]*net.IPNet, 0)

	cdirs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cdir := range cdirs {
		_, block, _ := net.ParseCIDR(cdir)
		blocks = append(blocks, block)
	}

	return blocks
}

func NetworkInterfaceHandler(w http.ResponseWriter, r *http.Request) {
	interfaces, err := NetworkInterfaces()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data, err := json.MarshalIndent(interfaces, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
