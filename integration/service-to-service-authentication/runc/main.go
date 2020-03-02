package main

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

func main() {
	logger, err := run.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	logger.Notice("Starting runc...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling HTTP request...")
		w.Write([]byte("Response from runc"))
	})

	log.Fatal(run.ListenAndServe(nil))
}
