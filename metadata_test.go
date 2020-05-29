package run

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testID               = "00bf4bf02db3546595153c211fec26b688516bf1b609f7d5e2f5d9ae17d1bcbaf6ce6c0c1b1168abf1ab255125e84e085336a36ae5715b0f95e7"
	testNumericProjectID = "123456789"
	testProjectID        = "test"
	testRegion           = "test"
)

const testAccessToken = `{
  "access_token": "ya29.AHES6ZRVmB7fkLtd1XTmq6mo0S1wqZZi3-Lh_s-6Uw7p8vtgSwg",
  "expires_in": 3484,
  "token_type": "Bearer"
}`

const (
	testIDToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.Et9HFtf9R3GEMA0IICOfFMVXY7kkTX1wr4qCyhIf58U"
)

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/computeMetadata/v1/instance/id" {
		fmt.Fprint(w, testID)
		return
	}

	if path == "/computeMetadata/v1/project/project-id" {
		fmt.Fprint(w, testProjectID)
		return
	}

	if path == "/computeMetadata/v1/project/numeric-project-id" {
		fmt.Fprint(w, testNumericProjectID)
		return
	}

	if path == "/computeMetadata/v1/instance/zone" {
		fmt.Fprint(w, fmt.Sprintf("projects/%s/zones/%s-1", testNumericProjectID, testRegion))
		return
	}

	if path == "/computeMetadata/v1/instance/region" {
		fmt.Fprint(w, testRegion)
		return
	}

	if path == "/computeMetadata/v1/instance/service-accounts/default/token" {
		fmt.Fprint(w, testAccessToken)
		return
	}

	if path == "/computeMetadata/v1/instance/service-accounts/default/identity" {
		fmt.Fprint(w, testIDToken)
		return
	}
}

func TestID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	id, err := ID()
	if err != nil {
		t.Error(err)
	}

	if id != testID {
		t.Errorf("got id %v, want %v", id, testID)
	}
}

func TestProjectID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	id, err := ProjectID()
	if err != nil {
		t.Error(err)
	}

	if id != testProjectID {
		t.Errorf("got project id %v, want %v", id, testProjectID)
	}
}

func TestProjectNumber(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	id, err := NumericProjectID()
	if err != nil {
		t.Error(err)
	}

	if id != testNumericProjectID {
		t.Errorf("got numeric project id %v, want %v", id, testNumericProjectID)
	}
}

func TestRegion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	region, err := Region()
	if err != nil {
		t.Error(err)
	}

	if region != testRegion {
		t.Errorf("got region %v, want %v", region, testRegion)
	}
}

func TestIDToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(metadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	token, err := IDToken("https://test-0123456789-ue.a.run.app")
	if err != nil {
		t.Error(err)
	}

	if token != testIDToken {
		t.Errorf("got token %v, want %v", token, testIDToken)
	}
}
