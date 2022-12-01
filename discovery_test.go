package run

import (
	"github.com/kelseyhightower/run/internal/gcptest"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEndpoints(t *testing.T) {
	ms := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	ss := httptest.NewServer(http.HandlerFunc(gcptest.ServiceDirectoryHandler))
	defer ss.Close()

	serviceDirectoryEndpoint = ss.URL

	endpoints, err := Endpoints("test", "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := []Endpoint{
		Endpoint{
			Name:    "test-10-0-0-1",
			Address: "10.0.0.1",
			Port:    8080,
		},
		Endpoint{
			Name:    "test-10-0-0-2",
			Address: "10.0.0.2",
			Port:    8080,
		},
	}

	if !reflect.DeepEqual(endpoints, expected) {
		t.Errorf("want %v, got %v", expected, endpoints)
	}
}
