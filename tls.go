package run

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	MTLSConfigurationPath = "mtls.json"
)

type MTLSConfig struct {
	AllowList     []string `json:"allow_list"`
	CACertificate string   `json:"ca"`
	Certificate   string   `json:"certificate"`
	Key           string   `json:"key"`
}

type MTLSConfigManager struct {
	AllowList   []string
	Certificate *tls.Certificate
	CertPool    *x509.CertPool

	config *MTLSConfig
}

func (m *MTLSConfigManager) LoadConfig() error {
	var config MTLSConfig

	data, err := os.ReadFile(MTLSConfigurationPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	m.config = &config

	certPEMBlock, err := base64.StdEncoding.DecodeString(m.config.Certificate)
	if err != nil {
		return err
	}

	keyPEMBlock, err := base64.StdEncoding.DecodeString(m.config.Key)
	if err != nil {
		return err
	}

	certificate, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return err
	}

	m.Certificate = &certificate

	caCertPEMBlock, err := base64.StdEncoding.DecodeString(m.config.CACertificate)
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCertPEMBlock)

	m.CertPool = certPool

	m.AllowList = config.AllowList

	return nil
}

func (m *MTLSConfigManager) GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	_ = info

	if m.config == nil {
		if err := m.LoadConfig(); err != nil {
			return nil, err
		}
	}

	return m.Certificate, nil
}

func (m *MTLSConfigManager) GetClientCertificate(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	_ = info

	if m.config == nil {
		if err := m.LoadConfig(); err != nil {
			return nil, err
		}
	}

	return m.Certificate, nil
}

func (m *MTLSConfigManager) VerifyPeerSPIFFECertificate() func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		if m.config == nil {
			if err := m.LoadConfig(); err != nil {
				return err
			}
		}

		var certs []*x509.Certificate

		for _, rawCert := range rawCerts {
			cert, err := x509.ParseCertificate(rawCert)
			if err != nil {
				return err
			}
			certs = append(certs, cert)
		}

		if len(certs) == 0 {
			return errors.New("no certificates found")
		}

		leaf := certs[0]

		spiffeID, err := spiffeIDFromCertificate(leaf)
		if err != nil {
			return err
		}

		_, err = leaf.Verify(x509.VerifyOptions{
			Roots:       m.CertPool,
			KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			CurrentTime: time.Now(),
		})

		if err != nil {
			return err
		}

		for _, id := range m.AllowList {
			if spiffeID == id {
				return nil
			}
		}

		return errors.New(fmt.Sprintf("SPIFFE ID rejected: %s", spiffeID))
	}
}

func spiffeIDFromCertificate(cert *x509.Certificate) (string, error) {
	if len(cert.URIs) == 0 {
		return "", errors.New("missing SPIFFE ID")
	}

	if len(cert.URIs) > 1 {
		return "", errors.New("more than one URI SAN found")
	}

	return cert.URIs[0].String(), nil
}
