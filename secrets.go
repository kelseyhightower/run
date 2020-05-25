package run

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	secretmanagerEndpoint = "https://secretmanager.googleapis.com/v1"
)

// SecretManagerPermissionError is returned when calls to the Secret
// Manager API fail to authenticate.
type SecretManagerPermissionError struct {
	Endpoint   string
	Name       string
	StatusCode int
}

func (e *SecretManagerPermissionError) Error() string {
	return "run/secret-manager: permission denied"
}

// SecretNotFoundError is returned when a secret version does not
// exist in the Secret Manager API.
type SecretNotFoundError struct {
	Endpoint string
	Name     string
}

func (e *SecretNotFoundError) Error() string {
	return "run: secret not found"
}

// SecretManagerError is returned when calls to the Secret Manager
// API fail for an unknown reason.
type SecretManagerError struct {
	Endpoint   string
	Name       string
	StatusCode int
}

func (e *SecretManagerError) Error() string {
	return "run: error retrieving secret"
}

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
func AccessSecretVersion(name, version string) (string, error) {
	return accessSecretVersion(name, version)
}

// AccessSecret returns the latest version of a Google Cloud Secret
// for the given name.
func AccessSecret(name string) (string, error) {
	return accessSecretVersion(name, "latest")
}

func accessSecretVersion(name, version string) (string, error) {
	if version == "" {
		version = "latest"
	}

	token, err := Token([]string{"https://www.googleapis.com/auth/cloud-platform"})
	if err != nil {
		return "", err
	}

	numericProjectID, err := NumericProjectID()
	if err != nil {
		return "", err
	}

	secretVersion := formatSecretVersion(numericProjectID, name, version)
	endpoint := fmt.Sprintf("%s/%s:access", secretmanagerEndpoint, secretVersion)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("User-Agent", userAgent)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	timeout := time.Duration(5) * time.Second
	httpClient := http.Client{Timeout: timeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	switch s := response.StatusCode; s {
	case 200:
		break
	case 401, 403:
		return "", &SecretManagerPermissionError{endpoint, name, s}
	case 404:
		return "", &SecretNotFoundError{endpoint, name}
	default:
		return "", &SecretManagerError{endpoint, name, s}
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var s SecretVersion
	err = json.Unmarshal(data, &s)
	if err != nil {
		return "", err
	}

	decodedString, err := base64.StdEncoding.DecodeString(s.Payload.Data)
	if err != nil {
		return "", err
	}

	return string(decodedString), nil
}
