package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/kelseyhightower/run"
)

type ServiceAuthenticationTestResults struct {
	Status string `json:"status"`
}

func serviceAuthenticationTestHandler(w http.ResponseWriter, r *http.Request) {
	run.Info(r, "Starting service authentication tests...")

	request, err := http.NewRequest("GET", "https://run-integration-backend", nil)
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	tr := &run.Transport{
		EnableServiceNameResolution: true,
	}

	httpClient := http.Client{Transport: tr}

	response, err := httpClient.Do(request)
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	result := ServiceAuthenticationTestResults{
		Status: string(responseData),
	}

	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		run.Error(r, err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(data)
}
