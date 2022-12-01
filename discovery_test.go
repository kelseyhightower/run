package run

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/kelseyhightower/run/internal/gcptest"
)

var testEndpoints = []Endpoint{
	{
		Name:    "test-10-0-0-1",
		Address: "10.0.0.1",
		Port:    8080,
	},
	{
		Name:    "test-10-0-0-2",
		Address: "10.0.0.2",
		Port:    8080,
	},
}

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

	expected := testEndpoints

	if !reflect.DeepEqual(endpoints, expected) {
		t.Errorf("want %v, got %v", expected, endpoints)
	}
}

var newRoundRobinLoadBalancerTests = []struct {
	namespace string
	name      string
	want      []Endpoint
}{
	{"test", "test", testEndpoints},
}

func TestNewRoundRobinLoadBalancer(t *testing.T) {
	ms := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	ss := httptest.NewServer(http.HandlerFunc(gcptest.ServiceDirectoryHandler))
	defer ss.Close()

	serviceDirectoryEndpoint = ss.URL

	for _, tt := range newRoundRobinLoadBalancerTests {
		lb, err := NewRoundRobinLoadBalancer(tt.namespace, tt.name)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		for i := 0; i <= len(tt.want)-1; i++ {
			endpoint := lb.Next()
			if !reflect.DeepEqual(endpoint, tt.want[i]) {
				t.Errorf("want %v, got %v", tt.want[i], endpoint)
			}
		}
	}
}
