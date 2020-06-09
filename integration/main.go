package main

import (
	"net/http"

	"github.com/kelseyhightower/run"
)

func main() {
	run.Notice("Starting run integration tests...")

	http.HandleFunc("/tests/env", envTestHandler)
	http.HandleFunc("/tests/metadata", metadataTestHandler)
	http.HandleFunc("/tests/secrets", secretsTestHandler)
	http.HandleFunc("/tests/service-authentication", serviceAuthenticationTestHandler)

	run.Fatal(run.ListenAndServe(nil))
}
