package run

import (
	"errors"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var cloudrunEndpoint = "https://%s-run.googleapis.com"

// Service represents a Cloud Run service.
type Service struct {
	Status ServiceStatus `json:"status"`
}

// ServiceStatus holds the current state of the Cloud Run service.
type ServiceStatus struct {
	// URL holds the url that will distribute traffic over the
	// provided traffic targets. It generally has the form
	// https://{route-hash}-{project-hash}-{cluster-level-suffix}.a.run.app
	URL string `json:"url"`

	// Similar to url, information on where the service is available on HTTP.
	Address ServiceAddresss `json:"address"`
}

type ServiceAddresss struct {
	URL string `json:"url"`
}

func regionalEndpoint(region string) string {
	return fmt.Sprintf(cloudrunEndpoint, region)
}

func getService(name, region, project string) (*Service, error) {
	var err error

	if region == "" {
		region, err = Region()
		if err != nil {
			return nil, err
		}
	}

	if project == "" {
		project, err = ProjectID()
		if err != nil {
			return nil, err
		}
	}

	endpoint := fmt.Sprintf("%s/apis/serving.knative.dev/v1/namespaces/%s/services/%s",
		regionalEndpoint(region), project, name)

	token, err := Token([]string{"https://www.googleapis.com/auth/cloud-platform"})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New("run: error resolving service name")
	}

	var service Service

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &service)
	if err != nil {
		return nil, err
	}

	return &service, nil
}
