package run

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransport(t *testing.T) {
	ms := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	var headers http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		return
	}))
	defer ts.Close()

	httpClient := &http.Client{Transport: &Transport{}}

	response, err := httpClient.Get(ts.URL)
	if err != nil {
		t.Error(err)
	}
	defer response.Body.Close()

	authHeader := headers.Get("Authorization")
	expectedAuthHeader := fmt.Sprintf("Bearer %s", testIDToken)

	if authHeader != expectedAuthHeader {
		t.Errorf("headers mismatch; want %s, got %s", expectedAuthHeader, authHeader)
	}
}

func TestTransportEnableServiceNameResolution(t *testing.T) {
	ms := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	headers := make(http.Header)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers.Set("x-test-result", "success")
	}))
	defer ts.Close()

	services := map[string]string{
		"test": ts.URL,
	}

	crs := httptest.NewServer(cloudrunServer(services))
	defer crs.Close()

	cloudrunEndpoint = crs.URL

	httpClient := &http.Client{
		Transport: &Transport{
			EnableServiceNameResolution: true,
		},
	}

	response, err := httpClient.Get("http://test")
	if err != nil {
		t.Error(err)
	}
	defer response.Body.Close()

	testHeader := headers.Get("x-test-result")
	if testHeader != "success" {
		t.Errorf("headers mismatch; want %s, got %s", "success", testHeader)
	}
}
