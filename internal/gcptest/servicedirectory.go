package gcptest

import (
	"fmt"
	"net/http"
)

var testEndpoints = `{
  "endpoints": [
    {
      "name": "test-10-0-0-1",
      "address": "10.0.0.1",
      "port": 8080
    },
    {
      "name": "test-10-0-0-2",
      "address": "10.0.0.2",
      "port": 8080
    }
  ]
}`

func ServiceDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/v1/projects/test/locations/test/namespaces/test/services/test/endpoints" {
		r.Header.Set("Content-Type", "application/json")
		fmt.Fprintf(w, testEndpoints)
		return
	}
}
