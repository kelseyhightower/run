package gcptest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Service represents a Cloud Run service.
type Service struct {
    Status ServiceStatus `json:"status"`
}

// ServiceStatus holds the current state of the Cloud Run service.
type ServiceStatus struct {
    // URL holds the url that will distribute traffic over the
    // provided traffic targets. It generally has the form
    // https://{route-hash}-{project-hash}-{cluster-level-suffix}.a.run.app
    URL string `json:"url"`

    // Similar to url, information on where the service is available on HTTP.
    Address ServiceAddresss `json:"address"`
}

type ServiceAddresss struct {
    URL string `json:"url"`
}

type cloudrunHandler struct {
	services map[string]string
}

func CloudrunServer(services map[string]string) http.Handler {
	return &cloudrunHandler{services}
}

func (h *cloudrunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := h.services["test"]

	s := Service{
		Status: ServiceStatus{
			URL: u,
			Address: ServiceAddresss{
				URL: u,
			},
		},
	}

	path := r.URL.Path

	if path == "/apis/serving.knative.dev/v1/namespaces/test/services/test" {
		data, err := json.Marshal(s)
		if err != nil {
			http.Error(w, "", 500)
			return
		}

		fmt.Fprint(w, string(data))
		return
	}

	if path == "/apis/serving.knative.dev/v1/namespaces/test/services/not-found" {
		http.NotFound(w, r)
		return
	}

	http.Error(w, "", 500)
}
