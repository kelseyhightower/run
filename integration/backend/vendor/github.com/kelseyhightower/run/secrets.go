package run

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	secretmanagerEndpoint = "https://secretmanager.googleapis.com/v1"
)

// ErrSecretPermissionDenied is returned when access to a secret is denied.
var ErrSecretPermissionDenied = errors.New("run: permission denied to named secret")

// ErrSecretUnauthorized is returned when calls to the Secret
// Manager API are unauthorized.
var ErrSecretUnauthorized = errors.New("run: secret manager unauthorized")

// ErrSecretNotFound is returned when a secret is not found.
var ErrSecretNotFound = errors.New("run: named secret not found")

// ErrSecretUnknownError is return when calls to the Secret Manager
// API return an unknown error.
var ErrSecretUnknownError = errors.New("run: unexpected error retrieving named secret")

// ErrSecretUnexpectedResponse is returned when calls to the Secret Manager
// API return an unexpected response.
type ErrSecretUnexpectedResponse struct {
	StatusCode int
	Err        error
}

func (e *ErrSecretUnexpectedResponse) Error() string {
	return "run: unexpected error retrieving named secret"
}

func (e *ErrSecretUnexpectedResponse) Unwrap() error { return e.Err }

// SecretVersion represents a Google Cloud Secret.
type SecretVersion struct {
	Name    string
	Payload SecretPayload `json:"payload"`
}

// SecretPayload holds the secret payload for a Google Cloud Secret.
type SecretPayload struct {
	// A base64-encoded string.
	Data string `json:"data"`
}

func formatSecretVersion(project, name, version string) string {
	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, name, version)
}

// AccessSecretVersion returns a Google Cloud Secret for the given
// secret name and version.
func AccessSecretVersion(name, version string) ([]byte, error) {
	return accessSecretVersion(name, version)
}

// AccessSecret returns the latest version of a Google Cloud Secret
// for the given name.
func AccessSecret(name string) ([]byte, error) {
	return accessSecretVersion(name, "latest")
}

func accessSecretVersion(name, version string) ([]byte, error) {
	if version == "" {
		version = "latest"
	}

	token, err := Token([]string{"https://www.googleapis.com/auth/cloud-platform"})
	if err != nil {
		return nil, err
	}

	numericProjectID, err := NumericProjectID()
	if err != nil {
		return nil, err
	}

	secretVersion := formatSecretVersion(numericProjectID, name, version)
	endpoint := fmt.Sprintf("%s/%s:access", secretmanagerEndpoint, secretVersion)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", userAgent)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

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
	case 401:
		return nil, ErrSecretUnauthorized
	case 403:
		return nil, ErrSecretPermissionDenied
	case 404:
		return nil, ErrSecretNotFound
	default:
		return nil, &ErrSecretUnexpectedResponse{s, ErrSecretUnknownError}
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var s SecretVersion
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}

	decodedData, err := base64.StdEncoding.DecodeString(s.Payload.Data)
	if err != nil {
		return nil, err
	}

	return decodedData, nil
}
