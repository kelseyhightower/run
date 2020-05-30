package run

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelseyhightower/run/internal/gcptest"
)

var accessSecretTests = []struct {
	name string
	want string
	err  error
}{
	{"foo", "Test", nil},
	{"bar", "", ErrSecretNotFound},
	{"denied", "", ErrSecretPermissionDenied},
	{"unauthorized", "", ErrSecretUnauthorized},
	{"unexpected", "", ErrSecretUnknownError},
}

func TestAccessSecret(t *testing.T) {
	ms := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	ss := httptest.NewServer(http.HandlerFunc(gcptest.SecretsHandler))
	defer ss.Close()

	secretmanagerEndpoint = ss.URL

	for _, tt := range accessSecretTests {
		secret, err := AccessSecret(tt.name)
		if !errors.Is(err, tt.err) {
			t.Errorf("unexpected error, want %q, got %q", tt.err, err)
		}

		if secret != tt.want {
			t.Errorf("want %s, got %s", tt.want, secret)
		}
	}
}
