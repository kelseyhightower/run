package main

import (
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/run"
)

type SecretsTestResults struct {
	Secret string `json:"secret"`
}

func secretsTestHandler(w http.ResponseWriter, r *http.Request) {
	run.Info(r, "Starting secrets tests...")

	secret, err := run.AccessSecret("run-secrets-integration-tests")
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	result := &SecretsTestResults{
		Secret: secret,
	}

	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}
