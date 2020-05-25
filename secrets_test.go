package run

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"testing"
)

var fooSecret = `{
  "name": "projects/123456789/secrets/foo/versions/1",
  "payload": {
    "data": "VGVzdA=="
  }
}`

func secretsHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/projects/123456789/secrets/unexpected/versions/latest:access" {
		http.Error(w, "", 500)
	}

	if path == "/projects/123456789/secrets/unauthorized/versions/latest:access" {
		http.Error(w, "", 401)
	}

	if path == "/projects/123456789/secrets/denied/versions/latest:access" {
		http.Error(w, "", 403)
	}

	if path == "/projects/123456789/secrets/bar/versions/latest:access" {
		http.NotFound(w, r)
	}

	if path == "/projects/123456789/secrets/foo/versions/latest:access" {
		fmt.Fprint(w, fooSecret)
		return
	}
}

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
	ms := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ms.Close()

	metadataEndpoint = ms.URL

	ss := httptest.NewServer(http.HandlerFunc(secretsHandler))
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
