package run

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	secretmanagerEndpoint = "https://secretmanager.googleapis.com/v1beta1"
)

type SecretVersion struct {
	Name    string
	Payload SecretPayload `json:"payload"`
}

type SecretPayload struct {
	// A base64-encoded string.
	Data string `json:"data"`
}

func formatSecretVersion(project, name, version string) string {
	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, name, version)
}

func AccessSecretVersion(name, version string) (string, error) {
	return accessSecretVersion(name, version)
}

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

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

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
