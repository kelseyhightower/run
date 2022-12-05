package run

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelseyhightower/run/internal/gcptest"
)

func TestNewTLSClientConfig(t *testing.T) {
	MTLSConfigurationPath = "testdata/server-mtls.json"

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		return
	})

	m := &MTLSConfigManager{}

	tt := httptest.NewUnstartedServer(nil)
	tt.TLS = &tls.Config{
		ClientAuth:            tls.RequireAnyClientCert,
		InsecureSkipVerify:    true,
		VerifyPeerCertificate: m.VerifyPeerSPIFFECertificate(),
		GetCertificate:        m.GetCertificate,
	}

	tt.StartTLS()
	defer tt.Close()

	// Remove the certificates added by the test server to force
	// the GetCertificate() method to be used.
	tt.TLS.Certificates = []tls.Certificate{}

	ms := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	httpClient := Client{
		EnableMutualTLSAuthentication: true,
	}

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/ping", tt.URL), nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	response, err := httpClient.Do(r)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.StatusCode != 200 {
		t.Errorf("want %v, got %v", 200, response.StatusCode)
	}
}
