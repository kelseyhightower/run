package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

var (
	backendServiceURL string
)

func main() {
	flag.StringVar(&backendServiceURL, "backend-service-url", "", "The backend service URL")
	flag.Parse()

	logger, err := run.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	logger.Notice("Starting frontend service...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling HTTP request...")
		request, err := http.NewRequest("GET", backendServiceURL, nil)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		tr := &run.Transport{
			EnableServiceNameResolution: true,
		}

		httpClient := http.Client{Transport: tr}

		response, err := httpClient.Do(request)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		defer response.Body.Close()

		w.Write([]byte(fmt.Sprintf("Backend response: %s", data)))
	})

	log.Fatal(run.ListenAndServe(nil))
}
