package run

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/kelseyhightower/run/metadata"
)

// Transport is an HTTP transport.
type Transport struct {
	tr http.RoundTripper
}

// RoundTrip is a round tripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	idToken, err := metadata.IDToken(audFromRequest(r))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))
	return t.tr.RoundTrip(req)
}

// audFromRequest extracts the Cloud Run service URL from an HTTP request.
func audFromRequest(r *http.Request) string {
	return fmt.Sprintf("%s://%s", r.URL.Scheme, r.URL.Hostname())
}

// ListenAndServe is serving.
func ListenAndServe(handler http.Handler) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := net.JoinHostPort("0.0.0.0", port)

	server := &http.Server{Addr: addr, Handler: handler}

	return server.ListenAndServe()
}
