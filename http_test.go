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
