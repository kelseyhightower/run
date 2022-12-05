package run

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelseyhightower/run/internal/gcptest"
)

func TestParseHostname(t *testing.T) {
	host := "ping.default.run.local"
	hostname, err := parseHostname(host)
	if err != nil {
		t.Error(err)
	}

	if hostname.Service != "ping" {
		t.Errorf("service name mismatch; want %v, got %v", "ping", hostname.Service)
	}
}

func TestTransport(t *testing.T) {
	ms := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	ss := httptest.NewServer(http.HandlerFunc(gcptest.ServiceDirectoryHandler))
	defer ss.Close()

	serviceDirectoryEndpoint = ss.URL

	var headers http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		return
	}))
	defer ts.Close()

	httpClient := &http.Client{
		Transport: &Transport{
			Base:      http.DefaultTransport,
			balancers: make(map[string]*RoundRobinLoadBalancer),
		},
	}

	response, err := httpClient.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}
	defer response.Body.Close()

	authHeader := headers.Get("Authorization")
	expectedAuthHeader := fmt.Sprintf("Bearer %s", gcptest.IDToken)

	if authHeader != expectedAuthHeader {
		t.Errorf("headers mismatch; want %s, got %s", expectedAuthHeader, authHeader)
	}
}

func TestTransportNameResolution(t *testing.T) {
	ms := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	ss := httptest.NewServer(http.HandlerFunc(gcptest.ServiceDirectoryHandler))
	defer ss.Close()

	serviceDirectoryEndpoint = ss.URL

	headers := make(http.Header)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers.Set("x-test-result", "success")
	}))
	defer ts.Close()

	httpClient := &http.Client{
		Transport: &Transport{
			Base:      http.DefaultTransport,
			balancers: make(map[string]*RoundRobinLoadBalancer),
		},
	}

	_, err := httpClient.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}

	testHeader := headers.Get("x-test-result")
	if testHeader != "success" {
		t.Errorf("headers mismatch; want %s, got %s", "success", testHeader)
	}
}
