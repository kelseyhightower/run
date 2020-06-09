package main

import (
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/run"
)

type MetadataTestResults struct {
	ID               string `json:"id"`
	NumericProjectID string `json:"numeric_project_id"`
	ProjectID        string `json:"project_id"`
	Region           string `json:"region"`
}

func metadataTestHandler(w http.ResponseWriter, r *http.Request) {
	run.Info(r, "Starting metadata tests...")

	id, err := run.ID()
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	numericProjectID, err := run.NumericProjectID()
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	projectID, err := run.ProjectID()
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	region, err := run.Region()
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	result := MetadataTestResults{
		ID:               id,
		NumericProjectID: numericProjectID,
		ProjectID:        projectID,
		Region:           region,
	}

	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}
