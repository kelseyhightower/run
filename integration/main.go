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

	logger.Notice("Starting integration app...")
	logger.Error("This is an error.")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r, "Handling request...")

		region, err := run.Region()
		if err != nil {
			logger.Error(err)
		}

		w.Write([]byte(region))
	})

	log.Fatal(run.ListenAndServe(nil))
}
