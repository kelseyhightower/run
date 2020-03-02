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
	runcServiceURL string
)

func main() {
	flag.StringVar(&runcServiceURL, "runc-service-url", "", "The runc service URL")
	flag.Parse()

	logger, err := run.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	logger.Notice("Starting rund...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Handling HTTP request...")
		request, err := http.NewRequest("GET", runcServiceURL, nil)
		if err != nil {
			logger.Error(err)
			http.Error(w, err.Error(), 500)
			return
		}

		response, err := run.HTTPClient.Do(request)
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
