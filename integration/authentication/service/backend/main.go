package main

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

func main() {
	logger := run.NewLogger()

	logger.Notice("Starting backend service...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling HTTP request...")
		w.Write([]byte("Response from backend"))
	})

	log.Fatal(run.ListenAndServe(nil))
}
