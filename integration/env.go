package main

import (
	"encoding/json"
	"net/http"

	"github.com/kelseyhightower/run"
)

type EnvTestResults struct {
	Port          string `json:"port"`
	Revision      string `json:"revision"`
	Configuration string `json:"configuration"`
	ServiceName   string `json:"service_name"`
}

func envTestHandler(w http.ResponseWriter, r *http.Request) {
	run.Info(r, "Starting env tests...")

	result := EnvTestResults{
		Port:          run.Port(),
		Revision:      run.Revision(),
		Configuration: run.Configuration(),
		ServiceName:   run.ServiceName(),
	}

	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}
