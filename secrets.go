package main

import (
	"encoding/base64"
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
	return fmt.Sprintf("name=projects/%s/secrets/%s/versions/%s", project, name, version)
}

func AccessSecret(name, version) (string, error) {
	token, err := Token()
	if err != nil {
		return "", err
	}

	numericProjectID, err := NumericProjectID()
	if err != nil {
		return "", err
	}

	secretVersion := formatSecretVersion(numericProjectID, name, version)
	endpoint := fmt.Sprintf("%s/%s", secretmanagerEndpoint, secretVersion)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

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
	err = json.Umarshal(data, &s)
	if err != nil {
		return "", err
	}

	decodedString, err := base64.StdEncoding.DecodeString(s.Data)
	if err != nil {
		return "", err
	}

	return string(decodedString), nil
}
