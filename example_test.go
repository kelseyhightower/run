package run_test

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

func ExampleTransport() {
	request, err := http.NewRequest("GET", "https://service-name", nil)
	if err != nil {
		log.Println(err)
		return
	}

	httpClient := &http.Client{
		Transport: &run.Transport{
			EnableServiceNameResolution: true,
		},
	}

	_, err = httpClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}
}
