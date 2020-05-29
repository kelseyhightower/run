package run

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type cloudrunHandler struct {
	services map[string]string
}

func cloudrunServer(services map[string]string) http.Handler {
	return &cloudrunHandler{services}
}

func (h *cloudrunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	testURL := h.services["test"]

	testService := Service{
		Status: ServiceStatus{
			URL: testURL,
			Address: ServiceAddresss{
				URL: testURL,
			},
		},
	}

	path := r.URL.Path

	if path == "/apis/serving.knative.dev/v1/namespaces/test/services/test" {
		data, err := json.Marshal(testService)
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
