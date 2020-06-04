package run

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kelseyhightower/run/internal/gcptest"
)

var metadataTests = []struct {
	name string
	want string
	err  error
}{
	{"id", gcptest.ID, nil},
	{"idtoken", gcptest.IDToken, nil},
	{"projectid", gcptest.ProjectID, nil},
	{"numericprojectid", gcptest.NumericProjectID, nil},
	{"region", gcptest.Region, nil},
	{"notfound", "", ErrMetadataNotFound},
	{"invalid", "", ErrMetadataInvalidRequest},
	{"unknown", "", ErrMetadataUnknownError},
}

func TestMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(gcptest.MetadataHandler))
	defer ts.Close()

	metadataEndpoint = ts.URL

	for _, tt := range metadataTests {
		var (
			err error
			v   string
		)

		switch tt.name {
		case "id":
			v, err = ID()
		case "idtoken":
			v, err = IDToken("https://test-0123456789-ue.a.run.app")
		case "projectid":
			v, err = ProjectID()
		case "numericprojectid":
			v, err = NumericProjectID()
		case "region":
			v, err = Region()
		default:
			v, err = errorMetadataRequest(tt.name)
		}

		if !errors.Is(err, tt.err) {
			t.Error(err)
		}

		if v != tt.want {
			t.Errorf("got id %v, want %v", v, tt.want)
		}
	}
}

func errorMetadataRequest(key string) (string, error) {
	endpoint := fmt.Sprintf("%s/computeMetadata/v1/%s", metadataEndpoint, key)
	v, err := metadataRequest(endpoint)
	return string(v), err
}
