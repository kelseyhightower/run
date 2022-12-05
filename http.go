package run

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	DefaultNamespace string
	DefaultRunDomain = "run.local"
)

type Client struct {
	// DisableInjectAuthorizationHeader, if true, prevents the Transport from
	// requesting an ID token from the metadata service, and populating
	// the Authorization request header when the http.Request contains
	// no existing Authorization value.
	DisableInjectAuthorizationHeader bool

	// EnableMutualTLSAuthentication, if true, enables mTLS on the client connection
	// based on a mTLS configuration bundle mounted at MTLSConfiguration
	// path.
	EnableMutualTLSAuthentication bool

	base          *http.Client
	baseClientSet bool
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.baseClientSet {
		return c.base.Do(req)
	}

	transport := &Transport{}

	if c.EnableMutualTLSAuthentication {
		m := &MTLSConfigManager{}

		transport.Base = &http.Transport{
			TLSClientConfig: &tls.Config{
				VerifyPeerCertificate: m.VerifyPeerSPIFFECertificate(),
				InsecureSkipVerify:    true,
				GetClientCertificate:  m.GetClientCertificate,
			},
		}
	}

	transport.DisableInjectAuthorizationHeader = c.DisableInjectAuthorizationHeader

	c.base = &http.Client{
		Transport: transport,
	}

	c.baseClientSet = true

	return c.base.Do(req)
}

// Transport is a http.RoundTripper that attaches ID tokens to all
// outgoing request.
type Transport struct {
	// Base optionally provides a http.RoundTripper that handles the
	// request. If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	// DisableInjectAuthorizationHeader, if true, prevents the Transport from
	// requesting an ID token from the metadata service, and populating
	// the Authorization request header when the http.Request contains
	// no existing Authorization value.
	DisableInjectAuthorizationHeader bool

	balancers map[string]*RoundRobinLoadBalancer
}

func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.Base == nil {
		t.Base = http.DefaultTransport
	}

	if !t.DisableInjectAuthorizationHeader {
		if r.Header.Get("Authorization") == "" {
			idToken, err := IDToken(audFromRequest(r))
			if err != nil {
				return nil, err
			}

			r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))
		}
	}

	hostname, err := parseHostname(r.Host)
	if err != nil {
		return t.Base.RoundTrip(r)
	}

	var loadBalancer *RoundRobinLoadBalancer

	serviceNamespace := fmt.Sprintf("%s.%s", hostname.Service, hostname.Namespace)
	if lb, ok := t.balancers[serviceNamespace]; ok {
		loadBalancer = lb
	} else {
		l, err := NewRoundRobinLoadBalancer(hostname.Service, hostname.Namespace)
		if err != nil {
			return nil, err
		}

		t.balancers[serviceNamespace] = l
		loadBalancer = l
	}

	endpoint := loadBalancer.Next()

	address := endpoint.Address
	port := endpoint.Port

	u, err := url.Parse(fmt.Sprintf("%s://%s:%d", r.URL.Scheme, address, port))
	if err != nil {
		return nil, err
	}

	r.Host = u.Host
	r.URL.Host = u.Host
	r.URL.Scheme = u.Scheme
	r.Header.Set("Host", u.Hostname())

	return t.Base.RoundTrip(r)
}

type Hostname struct {
	Domain    string
	Namespace string
	Service   string
}

var ErrInvalidHostname = errors.New("invalid hostname")

func parseHostname(host string) (*Hostname, error) {
	var hostname Hostname

	if strings.ContainsAny(host, ":") {
		return nil, ErrInvalidHostname
	}

	if ip := net.ParseIP(host); ip != nil {
		return nil, ErrInvalidHostname
	}

	ss := strings.Split(host, ".")

	switch len(ss) {
	case 0:
		return nil, ErrInvalidHostname
	case 1:
		hostname.Namespace = DefaultNamespace
		hostname.Service = ss[0]
	case 4:
		domain := fmt.Sprintf("%s.%s", ss[2], ss[3])
		if domain == DefaultRunDomain {
			hostname.Domain = domain
			hostname.Namespace = ss[1]
			hostname.Service = ss[0]
		}
	default:
		return nil, ErrInvalidHostname
	}

	return &hostname, nil
}

// audFromRequest extracts the Cloud Run service URL from an HTTP request.
func audFromRequest(r *http.Request) string {
	return fmt.Sprintf("%s://%s", r.URL.Scheme, r.URL.Hostname())
}

type Server struct {
	Addr     string
	Listener net.Listener
	Handler  http.Handler

	// Allowed holds a set of SPIFFE ID accepted by the HTTP client.
	Allowed []string
}

// ListenAndServe starts a http.Server with the given handler listening
// on the port defined by the PORT environment variable or "8080" if not
// set.
//
// ListenAndServe supports requests in HTTP/2 cleartext (h2c) format,
// because TLS is terminated by Cloud Run for all client requests including
// HTTP2.
//
// ListenAndServe traps the SIGINT and SIGTERM signals then gracefully
// shuts down the server without interrupting any active connections by
// calling the server's Shutdown method.
//
// ListenAndServe always returns a non-nil error; under normal conditions
// http.ErrServerClosed will be returned indicating a successful graceful
// shutdown.
func ListenAndServe(handler http.Handler) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if handler == nil {
		handler = http.DefaultServeMux
	}

	addr := net.JoinHostPort("0.0.0.0", port)

	h2s := &http2.Server{}
	server := &http.Server{Addr: addr, Handler: h2c.NewHandler(handler, h2s)}

	idleConnsClosed := make(chan struct{})
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan

		Notice("Received shutdown signal; waiting for active connections to close")

		if err := server.Shutdown(context.Background()); err != nil {
			Error("Error during server shutdown: %v", err)
		}

		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed

	Notice("Shutdown complete")

	return http.ErrServerClosed
}

func (s *Server) ListenAndServeMTLS() error {
	mTLSPort := os.Getenv("MTLS_PORT")
	if mTLSPort == "" {
		mTLSPort = "8443"
	}

	if s.Handler == nil {
		s.Handler = http.DefaultServeMux
	}

	if s.Addr == "" {
		s.Addr = net.JoinHostPort("0.0.0.0", mTLSPort)
	}

	m := &MTLSConfigManager{}

	h2s := &http2.Server{}
	server := &http.Server{
		Addr:    s.Addr,
		Handler: h2c.NewHandler(s.Handler, h2s),
		TLSConfig: &tls.Config{
			ClientAuth:            tls.RequireAnyClientCert,
			InsecureSkipVerify:    true,
			VerifyPeerCertificate: m.VerifyPeerSPIFFECertificate(),
			GetCertificate:        m.GetCertificate,
		},
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan

		Notice("Received shutdown signal; waiting for active connections to close")

		if err := server.Shutdown(context.Background()); err != nil {
			Error("Error during server shutdown: %v", err)
		}

		close(idleConnsClosed)
	}()

	if err := server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed

	Notice("Shutdown complete")

	return http.ErrServerClosed
}
