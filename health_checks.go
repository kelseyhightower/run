package run

import (
	"net/http"
)

// A Probe responds to health checks.
type Probe interface {
	Ready() bool
}

type httpProbeHandler struct {
	probe Probe
}

// Ready replies to the request with an HTTP 200 response.
func Ready(w http.ResponseWriter, r *http.Request) {
	Info(r, "HTTP startup probe succeeded")
	w.WriteHeader(200)
}

// Healthy replies to the request with an HTTP 200 response.
func Healthy(w http.ResponseWriter, r *http.Request) {
	Info(r, "HTTP liveness probe succeeded")
	w.WriteHeader(200)
}

// HTTPProbeHandler returns a request handler that calls the given probe and
// returns an HTTP 200 response if the probe Ready method returns true, or an
// HTTP 500 if false.
func HTTPProbeHandler(probe Probe) http.Handler {
	return &httpProbeHandler{probe: probe}
}

func (h *httpProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if h.probe.Ready() {
		w.WriteHeader(200)
		return
	}

	w.WriteHeader(500)
}
