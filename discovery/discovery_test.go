package discovery

import "testing"

var formatEndpointIDTests = []struct {
	name string
	ip   string
	want string
}{
	{"example", "10.0.0.1", "example-10-0-0-1"},
	{"example-service", "10.0.0.1", "example-service-10-0-0-1"},
}

func TestFormatEndpointID(t *testing.T) {
	for _, tt := range formatEndpointIDTests {
		endpointID := formatEndpointID(tt.name, tt.ip)
		if endpointID != tt.want {
			t.Errorf("string mismatch; want %v, got %v", tt.want, endpointID)
		}
	}
}
