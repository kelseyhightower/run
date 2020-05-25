package run

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var cloudrunEndpoint = "https://%s-run.googleapis.com"

// NameResolutionPermissionError is returned when service name
// resolution fails to authenticate with the Cloud Run API.
type NameResolutionPermissionError struct {
	Endpoint   string
	Name       string
	StatusCode int
}

func (e *NameResolutionPermissionError) Error() string {
	return "permission denied"
}

// ServiceNotFoundError is returned when a service does not exist
// in the Cloud Run API.
type ServiceNotFoundError struct {
	Endpoint string
	Name     string
}

func (e *ServiceNotFoundError) Error() string {
	return "not found"
}

// NameResolutionError is returned when service name resolution
// against the Cloud Run API fails for an unknown reason.
type NameResolutionError struct {
	Endpoint   string
	Name       string
	StatusCode int
}

func (e *NameResolutionError) Error() string {
	return fmt.Sprintf("error resolving service name %s", e.Name)
}

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

	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	switch s := response.StatusCode; s {
	case 200:
		break
	case 401:
		return nil, &NameResolutionPermissionError{endpoint, name, s}
	case 404:
		return nil, &ServiceNotFoundError{endpoint, name}
	default:
		return nil, &NameResolutionError{endpoint, name, s}
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
