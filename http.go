package run

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

// HTTPClient is an http.Client
var HTTPClient = &http.Client{
	Transport: &Transport{},
}

// Transport is an http.RoundTripper that attaches ID tokens to all
// all outgoing request.
type Transport struct {
	tr http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	idToken, err := IDToken(audFromRequest(req))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))
	if t.tr == nil {
		t.tr = http.DefaultTransport
	}
	return t.tr.RoundTrip(req)
}

// audFromRequest extracts the Cloud Run service URL from an HTTP request.
func audFromRequest(r *http.Request) string {
	return fmt.Sprintf("%s://%s", r.URL.Scheme, r.URL.Hostname())
}

// ListenAndServe starts an HTTP server with a given handler listening
// on the port defined by the PORT environment variable or "8080" if not
// set.
func ListenAndServe(handler http.Handler) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := net.JoinHostPort("0.0.0.0", port)

	server := &http.Server{Addr: addr, Handler: handler}

	return server.ListenAndServe()
}
