package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

type Result struct {
	Output string
}

func main() {
	logger, err := run.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	logger.Notice("Starting secrets integration tests...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		secret, err := run.AccessSecret("run-secrets-integration-tests")
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		result := &Result{
			Output: secret,
		}

		data, err := json.MarshalIndent(result, "", " ")
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(data)
	})

	log.Fatal(run.ListenAndServe(nil))
}
