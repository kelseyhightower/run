package discovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kelseyhightower/run"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var serviceEndpoint = "https://servicedirectory.googleapis.com"

type Service struct {
	Name string
}

type Endpoint struct {
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	Port        int               `json:"port"`
	Annotations map[string]string `json:"annotations"`
}

func RegisterEndpoint(namespace string) error {
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}
	token, err := run.Token(scopes)
	if err != nil {
		return err
	}

	c := http.Client{
		Timeout: time.Second * 10,
	}

	basePath, err := formatEndpointBasePath(namespace)
	if err != nil {
		return err
	}

	endpointID, err := generateEndpointID()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/%s?endpointId=%s", serviceEndpoint, basePath, endpointID)

	ip, err := run.IPAddress()
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(run.Port())
	if err != nil {
		return err
	}

	instanceID, err := run.ID()
	if err != nil {
		return err
	}

	annotations := make(map[string]string)
	annotations["instance_id"] = instanceID

	ep := Endpoint{
		Address:     ip,
		Port:        port,
		Annotations: annotations,
	}

	data, err := json.Marshal(ep)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(data)

	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		run.Log("Error", string(data))
		return errors.New(fmt.Sprintf("run: non 200 response when registering endpoint: %s", response.Status))
	}

	return nil
}

func DeregisterEndpoint(namespace string) error {
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}
	token, err := run.Token(scopes)
	if err != nil {
		return err
	}

	basePath, err := formatEndpointBasePath(namespace)
	if err != nil {
		return err
	}

	endpointID, err := generateEndpointID()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v1/%s/%s", serviceEndpoint, basePath, endpointID)

	c := http.Client{
		Timeout: time.Second * 10,
	}

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		run.Log("Error", string(data))
		return errors.New(fmt.Sprintf("run: non 200 response when deregistering endpoint: %s", response.Status))
	}

	return nil
}

func formatEndpointBasePath(namespace string) (string, error) {
	serviceName := run.ServiceName()
	region, err := run.Region()
	if err != nil {
		return "", err
	}

	projectID, err := run.ProjectID()
	if err != nil {
		return "", err
	}

	s := fmt.Sprintf("projects/%s/locations/%s/namespaces/%s/services/%s/endpoints",
		projectID, region, namespace, serviceName)

	return s, nil
}

func generateEndpointID() (string, error) {
	serviceName := run.ServiceName()

	ip, err := run.IPAddress()
	if err != nil {
		return "", err
	}

	dashedIPAddress := strings.ReplaceAll(ip, ".", "-")

	id := fmt.Sprintf("%s-%s", serviceName, dashedIPAddress)

	return id, nil
}
