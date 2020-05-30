package run

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelseyhightower/run/internal/gcptest"
)

func TestID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	id, err := ID()
	if err != nil {
		t.Error(err)
	}

	if id != gcptest.ID {
		t.Errorf("got id %v, want %v", id, gcptest.ID)
	}
}

func TestProjectID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	id, err := ProjectID()
	if err != nil {
		t.Error(err)
	}

	if id != gcptest.ProjectID {
		t.Errorf("got project id %v, want %v", id, gcptest.ProjectID)
	}
}

func TestProjectNumber(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	id, err := NumericProjectID()
	if err != nil {
		t.Error(err)
	}

	if id != gcptest.NumericProjectID {
		t.Errorf("got numeric project id %v, want %v", id, gcptest.NumericProjectID)
	}
}

func TestRegion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	region, err := Region()
	if err != nil {
		t.Error(err)
	}

	if region != gcptest.Region {
		t.Errorf("got region %v, want %v", region, gcptest.Region)
	}
}

func TestIDToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	token, err := IDToken("https://test-0123456789-ue.a.run.app")
	if err != nil {
		t.Error(err)
	}

	if token != gcptest.IDToken {
		t.Errorf("got token %v, want %v", token, gcptest.IDToken)
	}
}
