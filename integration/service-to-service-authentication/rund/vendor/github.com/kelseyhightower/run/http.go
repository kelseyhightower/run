package run

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// HTTPClient is an http.Client
var HTTPClient = &http.Client{
	Transport: &Transport{},
}

// Transport is an http.RoundTripper that attaches ID tokens to all
// all outgoing request.
type Transport struct {
	// DisableServiceNameResolution, if true, prevents the resolution
	// of service names using the Cloud Run API.
	//
	// When true, HTTP requests are modified by replacing the original
	// HTTP target URL with the service URL from the Cloud Run service
	// with a matching name in the same region as the caller.
	//
	// Examples:
	//
	//   http://service => https://service-6bn2iswfgq-ue.a.run.app
	//   https://service => https://service-6bn2iswfgq-ue.a.run.app
	//
	// Service accounts must have the roles/run.viewer IAM permission
	// to resolve service names using the Cloud Run API.
	DisableServiceNameResolution bool

	tr http.RoundTripper
}

func resolveServiceName(r *http.Request) error {
	var (
		serviceName string
		region      string
		project     string
	)

	parts := strings.Split(r.URL.Host, ".")

	switch n := len(parts); {
	case n > 1:
		return nil
	case n == 0:
		return nil
	case n == 1:
		serviceName = parts[0]
	}

	service, err := getService(serviceName, region, project)
	if err != nil {
		return err
	}

	u, err := url.Parse(service.Status.Address.URL)
	if err != nil {
		return err
	}

	r.Host = u.Host
	r.URL.Host = u.Host
	r.URL.Scheme = u.Scheme
	r.Header.Set("Host", u.Hostname())

	return nil
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.DisableServiceNameResolution {
		if err := resolveServiceName(req); err != nil {
			return nil, err
		}
	}

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
