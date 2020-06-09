package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

var rmu sync.Mutex

var (
	runtimeID               string
	runtimeProjectID        string
	runtimeRegion           string
	runtimeNumericProjectID string
)

var metadataEndpoint = "http://metadata.google.internal"

// ErrMetadataNotFound is returned when a metadata key is not found.
var ErrMetadataNotFound = errors.New("run: metadata key not found")

// ErrMetadataInvalidRequest is returned when a metadata request is invalid.
var ErrMetadataInvalidRequest = errors.New("run: invalid metadata request")

// ErrMetadataUnknownError is return when calls to the metadata server
// return an unknown error.
var ErrMetadataUnknownError = errors.New("run: unexpected error retrieving metadata key")

// ErrMetadataUnexpectedResponse is returned when calls to the metadata server
// return an unexpected response.
type ErrMetadataUnexpectedResponse struct {
	StatusCode int
	Err        error
}

func (e *ErrMetadataUnexpectedResponse) Error() string {
	return "run: unexpected error retrieving metadata key"
}

func (e *ErrMetadataUnexpectedResponse) Unwrap() error { return e.Err }

// AccessToken holds a GCP access token.
type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// ProjectID returns the active project ID from the metadata service.
func ProjectID() (string, error) {
	rmu.Lock()
	defer rmu.Unlock()

	if runtimeProjectID != "" {
		return runtimeProjectID, nil
	}

	endpoint := fmt.Sprintf("%s/computeMetadata/v1/project/project-id", metadataEndpoint)

	data, err := metadataRequest(endpoint)
	if err != nil {
		return "", err
	}

	runtimeProjectID = string(data)
	return runtimeProjectID, nil
}

// NumericProjectID returns the active project ID from the metadata service.
func NumericProjectID() (string, error) {
	rmu.Lock()
	defer rmu.Unlock()

	if runtimeNumericProjectID != "" {
		return runtimeNumericProjectID, nil
	}

	endpoint := fmt.Sprintf("%s/computeMetadata/v1/project/numeric-project-id", metadataEndpoint)

	data, err := metadataRequest(endpoint)
	if err != nil {
		return "", err
	}

	runtimeNumericProjectID = string(data)
	return runtimeNumericProjectID, nil
}

// Token returns the default service account token.
func Token(scopes []string) (*AccessToken, error) {
	s := strings.Join(scopes, ",")

	endpoint := fmt.Sprintf("%s/computeMetadata/v1/instance/service-accounts/default/token?scopes=%s", metadataEndpoint, s)
	data, err := metadataRequest(endpoint)
	if err != nil {
		return nil, err
	}

	var accessToken AccessToken
	err = json.Unmarshal(data, &accessToken)
	if err != nil {
		return nil, fmt.Errorf("run/metadata: error retrieving access token: %v", err)
	}

	return &accessToken, nil
}

// IDToken returns an id token based on the service url.
func IDToken(serviceURL string) (string, error) {
	endpoint := fmt.Sprintf("%s/computeMetadata/v1/instance/service-accounts/default/identity?audience=%s", metadataEndpoint, serviceURL)

	idToken, err := metadataRequest(endpoint)
	if err != nil {
		return "", fmt.Errorf("metadata.Get: failed to query id_token: %w", err)
	}
	return string(idToken), nil
}

// Region returns the name of the Cloud Run region.
func Region() (string, error) {
	rmu.Lock()
	defer rmu.Unlock()

	if runtimeRegion != "" {
		return runtimeRegion, nil
	}

	endpoint := fmt.Sprintf("%s/computeMetadata/v1/instance/region", metadataEndpoint)

	data, err := metadataRequest(endpoint)
	if err != nil {
		return "", err
	}

	runtimeRegion = path.Base(string(data))
	return runtimeRegion, nil
}

// ID returns the unique identifier of the container instance.
func ID() (string, error) {
	rmu.Lock()
	defer rmu.Unlock()

	if runtimeID != "" {
		return runtimeID, nil
	}

	endpoint := fmt.Sprintf("%s/computeMetadata/v1/instance/id", metadataEndpoint)

	data, err := metadataRequest(endpoint)
	if err != nil {
		return "", err
	}

	runtimeID = string(data)
	return runtimeID, nil
}

func metadataRequest(endpoint string) ([]byte, error) {
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", userAgent)
	request.Header.Add("Metadata-Flavor", "Google")

	timeout := time.Duration(5) * time.Second
	httpClient := http.Client{Timeout: timeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	switch s := response.StatusCode; s {
	case 200:
		break
	case 400:
		return nil, ErrMetadataInvalidRequest
	case 404:
		return nil, ErrMetadataNotFound
	default:
		return nil, &ErrMetadataUnexpectedResponse{s, ErrMetadataUnknownError}
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
