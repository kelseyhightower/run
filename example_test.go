package run_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kelseyhightower/run"
)

func ExampleAccessSecret() {
	secret, err := run.AccessSecret("apikey")
	if err != nil {
		log.Println(err)
		return
	}

	_ = secret
}

func ExampleAccessSecretVersion() {
	secret, err := run.AccessSecretVersion("apikey", "1")
	if err != nil {
		log.Println(err)
		return
	}

	_ = secret
}

func ExampleIDToken() {
	serviceURL := "https://example-6bn2iswfgq-uw.a.run.app"

	request, err := http.NewRequest(http.MethodGet, serviceURL, nil)
	if err != nil {
		log.Println(err)
		return
	}

	idToken, err := run.IDToken(serviceURL)
	if err != nil {
		log.Println(err)
		return
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))
}

func ExampleToken() {
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}

	project, err := run.ProjectID()
	if err != nil {
		log.Println(err)
		return
	}

	endpoint := fmt.Sprintf("https://cloudbuild.googleapis.com/v1/projects/%s/builds", project)

	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Println(err)
		return
	}

	token, err := run.Token(scopes)
	if err != nil {
		log.Println(err)
		return
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
}

func ExampleLogger() {
	logger := run.NewLogger()

	logger.Notice("Starting example service...")
}

func ExampleLogger_defaultLogger() {
	run.Notice("Starting example service...")
}

func ExampleLogger_logCorrelation() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Pass in the *http.Request as the first argument to correlate
		// container logs with request logs.
		run.Info(r, "Handling request...")
	})
}

func ExampleTransport() {
	client := &http.Client{Transport: &run.Transport{}}

	response, err := client.Get("https://example-6bn2iswfgq-uw.a.run.app")
	if err != nil {
		log.Println(err)
		return
	}

	defer response.Body.Close()
}

func ExampleTransport_serviceNameResolution() {
	client := &http.Client{
		Transport: &run.Transport{},
	}

	response, err := client.Get("https://service-name")
	if err != nil {
		log.Println(err)
		return
	}

	defer response.Body.Close()
}

func ExampleListenAndServe() {
	if err := run.ListenAndServe(nil); err != http.ErrServerClosed {
		run.Fatal(err)
	}
}
