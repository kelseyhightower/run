package main

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
	"github.com/kelseyhightower/run/metadata"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		region, err := metadata.Region()
		if err != nil {
			run.LogError(err)
		}

		w.Write([]byte(region))
	})

	log.Fatal(run.ListenAndServe(nil))
}
