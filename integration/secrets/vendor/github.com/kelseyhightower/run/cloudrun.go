package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var cloudrunEndpoint = "https://%s-run.googleapis.com"

// ErrNameResolutionPermissionDenied is returned when access to the
// Cloud Run API is denied.
var ErrNameResolutionPermissionDenied = errors.New("run: permission denied to named service")

// ErrNameResolutionUnauthorized is returned when calls to the Cloud
// Run API are unauthorized.
var ErrNameResolutionUnauthorized = errors.New("run: cloud run api unauthorized")

// ErrServiceNotFound is returned when a service is not found..
var ErrServiceNotFound = errors.New("run: named service not found")

// ErrNameResolutionUnknownError is return when calls to the Cloud Run
// API return an unknown error.
var ErrNameResolutionUnknownError = errors.New("run: unexpected error retrieving named service")

// ErrNameResolutionUnexpectedResponse is returned when calls to the Cloud Run
// API return an unexpected response.
type ErrNameResolutionUnexpectedResponse struct {
	StatusCode int
	Err        error
}

func (e *ErrNameResolutionUnexpectedResponse) Error() string {
	return "run: unexpected error retrieving named service"
}

func (e *ErrNameResolutionUnexpectedResponse) Unwrap() error { return e.Err }

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
	if region == "test" {
		return cloudrunEndpoint
	}
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

	request.Header.Set("User-Agent", userAgent)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	timeout := time.Duration(5) * time.Second
	httpClient := &http.Client{Timeout: timeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	switch s := response.StatusCode; s {
	case 200:
		break
	case 401:
		return nil, ErrNameResolutionUnauthorized
	case 403:
		return nil, ErrNameResolutionPermissionDenied
	case 404:
		return nil, ErrServiceNotFound
	default:
		return nil, &ErrNameResolutionUnexpectedResponse{s, ErrNameResolutionUnknownError}
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
