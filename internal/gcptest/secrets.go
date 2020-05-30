package gcptest

import (
	"fmt"
	"net/http"
)

const fooSecret = `{
  "name": "projects/123456789/secrets/foo/versions/1",
  "payload": {
    "data": "VGVzdA=="
  }
}`

func SecretsHandler(w http.ResponseWriter, r *http.Request) {
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
