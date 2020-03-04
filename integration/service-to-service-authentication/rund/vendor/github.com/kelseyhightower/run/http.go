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
	tr http.RoundTripper
}

func expandServiceURL(r *http.Request) error {
	var (
		serviceName string
		region      string
		project     string
	)

	parts := strings.Split(r.URL.Host, ".")

	switch n := len(parts); {
	case n > 2:
		return nil
	case n == 0:
		return nil
	case n == 1:
		serviceName = parts[0]
	case n == 2:
		region = parts[1]
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
	if err := expandServiceURL(req); err != nil {
		return nil, err
	}

	fmt.Printf("expanded url: %s\n", req.URL)
	fmt.Printf("host header: %s\n", req.Header.Get("Host"))

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
