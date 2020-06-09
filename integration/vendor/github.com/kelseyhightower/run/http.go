package run

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var serviceCache *cache

func init() {
	data := make(map[string]string)
	serviceCache = &cache{data}
}

// Transport is an http.RoundTripper that attaches ID tokens to all
// all outgoing request.
type Transport struct {
	// Base optionally provides an http.RoundTripper that handles the
	// request. If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	// EnableServiceNameResolution, if true, enables the resolution
	// of service names using the Cloud Run API.
	//
	// When true, HTTP requests are modified by replacing the original
	// HTTP target URL with the URL from the named Cloud Run service
	// in the same region as the caller.
	//
	// Examples:
	//
	//   http://service => https://service-6bn2iswfgq-ue.a.run.app
	//   https://service => https://service-6bn2iswfgq-ue.a.run.app
	//
	// Service accounts must have the roles/run.viewer IAM permission
	// to resolve service names using the Cloud Run API.
	EnableServiceNameResolution bool
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

	var u *url.URL
	endpoint := serviceCache.Get(serviceName)
	if endpoint == "" {
		service, err := getService(serviceName, region, project)
		if err != nil {
			return fmt.Errorf("run: error resolving service name: %w", err)
		}

		endpoint = service.Status.Address.URL
		serviceCache.Set(serviceName, endpoint)
	}

	u, err := url.Parse(endpoint)
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
	if t.EnableServiceNameResolution {
		if err := resolveServiceName(req); err != nil {
			return nil, err
		}
	}

	idToken, err := IDToken(audFromRequest(req))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))
	if t.Base == nil {
		t.Base = http.DefaultTransport
	}

	return t.Base.RoundTrip(req)
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
