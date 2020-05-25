package run_test

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

func ExampleTransport() {
	client := &http.Client{Transport: &run.Transport{}}

	_, err := client.Get("https://example-6bn2iswfgq-uw.a.run.app")
	if err != nil {
		log.Println(err)
		return
	}
}

func ExampleTransport_serviceNameResolution() {
	client := &http.Client{
		Transport: &run.Transport{
			EnableServiceNameResolution: true,
		},
	}

	_, err := client.Get("https://service-name")
	if err != nil {
		log.Println(err)
		return
	}
}
