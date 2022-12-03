package run

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var serviceDirectoryEndpoint = "https://servicedirectory.googleapis.com"

type Endpoint struct {
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Network     string            `json:"network,omitempty"`
	UID         string            `json:"uid,omitempty"`
}

type ListEndpoints struct {
	Endpoints []Endpoint `json:"endpoints"`
}

type LoadBalancer interface {
	Next() Endpoint
	RefreshEndpoints()
}

type RoundRobinLoadBalancer struct {
	name      string
	namespace string
	endpoints []Endpoint
	current   int
}

func NewRoundRobinLoadBalancer(name, namespace string) (*RoundRobinLoadBalancer, error) {
	endpoints, err := Endpoints(name, namespace)
	if err != nil {
		return nil, err
	}

	loadBalancer := &RoundRobinLoadBalancer{name, namespace, endpoints, 0}

	go loadBalancer.RefreshEndpoints()

	return loadBalancer, nil
}

func (lb *RoundRobinLoadBalancer) Next() Endpoint {
	endpoint := lb.endpoints[lb.current]
	lb.current++

	if lb.current >= len(lb.endpoints) {
		lb.current = 0
	}
	return endpoint
}

func (lb *RoundRobinLoadBalancer) RefreshEndpoints() {
	for {
		time.Sleep(time.Second * 10)
		endpoints, err := Endpoints(lb.name, lb.namespace)
		if err != nil {
			Log("Error", err.Error())
			continue
		}
		lb.endpoints = endpoints
		lb.current = 0
	}
}

func Endpoints(name, namespace string) ([]Endpoint, error) {
	var listEndpoints ListEndpoints

	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}
	token, err := Token(scopes)
	if err != nil {
		return nil, err
	}

	basePath, err := formatEndpointBasePath(name, namespace)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/%s", serviceDirectoryEndpoint, basePath)

	c := http.Client{
		Timeout: time.Second * 10,
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		Log("Error", string(data))
		return nil, errors.New(fmt.Sprintf("run: non 200 response when retrieving endpoints: %s", response.Status))
	}

	if err := json.Unmarshal(data, &listEndpoints); err != nil {
		return nil, err
	}

	return listEndpoints.Endpoints, nil
}

func RegisterEndpoint(namespace string) error {
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}
	token, err := Token(scopes)
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint: %s", err))
		return err
	}

	c := http.Client{
		Timeout: time.Second * 10,
	}

	basePath, err := formatEndpointBasePath("", namespace)
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. Error formating endpoint base path: %s", err))
		return err
	}

	endpointID, err := generateEndpointID()
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. Error generating endpoint ID: %s", err))
		return err
	}

	url := fmt.Sprintf("%s/v1/%s?endpointId=%s", serviceDirectoryEndpoint, basePath, endpointID)

	ip, err := IPAddress()
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. Error getting IP address: %s", err))
		return err
	}

	port, err := strconv.Atoi(Port())
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. Error converting instance port: %s", err))
		return err
	}

	instanceID, err := ID()
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. Error retrieving instance ID: %s", err))
		return err
	}

	annotations := make(map[string]string)
	annotations["instance_id"] = instanceID

	ep := Endpoint{
		Address:     ip,
		Port:        port,
		Annotations: annotations,
	}

	data, err := json.Marshal(ep)
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. Error serializing endpoint request object: %s", err))
		return err
	}

	body := bytes.NewBuffer(data)

	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. Error create endpoint HTTP request: %s", err))
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.Do(request)
	if err != nil {
		Log("Error", fmt.Sprintf("Unable to register endpoint. HTTP request failed: %s", err))
		return err
	}

	if response.StatusCode != 200 {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			Log("Error", fmt.Sprintf("Unable to register endpoint. Error reading HTTP response: %s", err))
			return err
		}

		Log("Error", fmt.Sprintf("Unable to register endpoint. HTTP request failed: %s", string(data)))
		return errors.New(fmt.Sprintf("run: non 200 response when registering endpoint: %s", response.Status))
	}

	Log("Info", fmt.Sprintf("Successfully registered endpoint: %s", endpointID))
	return nil
}

func DeregisterEndpoint(namespace string) error {
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}
	token, err := Token(scopes)
	if err != nil {
		return err
	}

	basePath, err := formatEndpointBasePath("", namespace)
	if err != nil {
		return err
	}

	endpointID, err := generateEndpointID()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/%s/%s", serviceDirectoryEndpoint, basePath, endpointID)

	c := http.Client{
		Timeout: time.Second * 10,
	}

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		Log("Error", string(data))
		return errors.New(fmt.Sprintf("run: non 200 response when deregistering endpoint: %s", response.Status))
	}

	return nil
}

func formatEndpointBasePath(name, namespace string) (string, error) {
	if name == "" {
		name = ServiceName()
	}
	region, err := Region()
	if err != nil {
		return "", err
	}

	projectID, err := ProjectID()
	if err != nil {
		return "", err
	}

	s := fmt.Sprintf("projects/%s/locations/%s/namespaces/%s/services/%s/endpoints",
		projectID, region, namespace, name)

	return s, nil
}

func generateEndpointID() (string, error) {
	serviceName := ServiceName()

	ip, err := IPAddress()
	if err != nil {
		return "", err
	}

	dashedIPAddress := strings.ReplaceAll(ip, ".", "-")

	id := fmt.Sprintf("%s-%s", serviceName, dashedIPAddress)

	return id, nil
}
