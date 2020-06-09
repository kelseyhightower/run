package main

import (
	"net/http"

	"github.com/kelseyhightower/run"
)

func main() {
	run.Notice("Starting integration backend service...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		run.Info(r, "Handling HTTP request...")
		w.Write([]byte("SUCCESS"))
	})

	run.Fatal(run.ListenAndServe(nil))
}
