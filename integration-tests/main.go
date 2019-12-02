package main

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
	"github.com/kelseyhightower/run/metadata"
)

func main() {
	logger := run.NewLogger("integration")
	logger.Notice("Starting integration app...")
	logger.Error("This is an error.")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r, "Handling request...")

		region, err := metadata.Region()
		if err != nil {
			logger.Error(err)
		}

		w.Write([]byte(region))
	})

	log.Fatal(run.ListenAndServe(nil))
}
