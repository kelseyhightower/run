package metadata

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"
)

func httpRequest(endpoint string) ([]byte, error) {
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Metadata-Flavor", "Google")

	timeout := time.Duration(3) * time.Second
	client := http.Client{Timeout: timeout}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	return data, nil
}

// IDToken returns an id token based on the service url.
func IDToken(serviceURL string) (string, error) {
	endpoint := fmt.Sprintf("http://metadata.google.internal/instance/service-accounts/default/identity?audience=%s", serviceURL)

	idToken, err := httpRequest(endpoint)
	if err != nil {
		return "", fmt.Errorf("metadata.Get: failed to query id_token: %+v", err)
	}
	return string(idToken), nil
}

// Region returns the cloud run region.
// https://cloud.google.com/run/docs/reference/container-contract
func Region() (string, error) {
	endpoint := "http://metadata.google.internal/computeMetadata/v1/instance/zone"

	data, err := httpRequest(endpoint)
	if err != nil {
		return "", err
	}

	region := strings.TrimSuffix(path.Base(string(data)), "-1")
	return region, nil
}
