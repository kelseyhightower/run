package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

type Result struct {
	ID               string `json:"id"`
	NumericProjectID string `json:"numeric_project_id"`
	ProjectID        string `json:"project_id"`
	Region           string `json:"region"`
}

func main() {
	logger, err := run.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	logger.Notice("Starting metadata integration tests...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id, err := run.ID()
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		numericProjectID, err := run.NumericProjectID()
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		projectID, err := run.ProjectID()
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		region, err := run.Region()
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		result := Result {
			ID:               id,
			NumericProjectID: numericProjectID,
			ProjectID:        projectID,
			Region:           region,
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
