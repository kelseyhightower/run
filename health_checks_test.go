package run

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReady(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	Ready(responseRecorder, request)
	response := responseRecorder.Result()

	if response.StatusCode != 200 {
		t.Errorf("status code mismatch; want %v, got %v", 200, response.StatusCode)
	}
}

func TestHealthy(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	Healthy(responseRecorder, request)
	response := responseRecorder.Result()

	if response.StatusCode != 200 {
		t.Errorf("status code mismatch; want %v, got %v", 200, response.StatusCode)
	}
}

type failProbe struct{}

func (p failProbe) Ready() bool {
	return false
}

type successProbe struct{}

func (p successProbe) Ready() bool {
	return true
}

var httpProbeHandlerTests = []struct {
	probe Probe
	want  int
}{
	{failProbe{}, 500},
	{successProbe{}, 200},
}

func TestHTTPProbeHandler(t *testing.T) {
	for _, tt := range httpProbeHandlerTests {
		responseRecorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/health", nil)

		HTTPProbeHandler(tt.probe).ServeHTTP(responseRecorder, request)
		response := responseRecorder.Result()
		if response.StatusCode != tt.want {
			t.Errorf("status code mismatch; want %v, got %v", tt.want, response.StatusCode)
		}
	}
}
