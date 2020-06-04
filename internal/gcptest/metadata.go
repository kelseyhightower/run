package gcptest

import (
	"fmt"
	"net/http"
)

const (
	ID               = "00bf4bf02db3546595153c211fec26b688516bf1b609f7d5e2f5d9ae17d1bcbaf6ce6c0c1b1168abf1ab255125e84e085336a36ae5715b0f95e7"
	NumericProjectID = "123456789"
	ProjectID        = "test"
	Region           = "test"
)

const AccessToken = `{
  "access_token": "ya29.AHES6ZRVmB7fkLtd1XTmq6mo0S1wqZZi3-Lh_s-6Uw7p8vtgSwg",
  "expires_in": 3484,
  "token_type": "Bearer"
}`

const (
	IDToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.Et9HFtf9R3GEMA0IICOfFMVXY7kkTX1wr4qCyhIf58U"
)

func MetadataHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/computeMetadata/v1/instance/id" {
		fmt.Fprint(w, ID)
		return
	}

	if path == "/computeMetadata/v1/notfound" {
		http.NotFound(w, r)
		return
	}

	if path == "/computeMetadata/v1/invalid" {
        http.Error(w, "", 400)
        return
    }

	if path == "/computeMetadata/v1/unknown" {
        http.Error(w, "", 500)
        return
    }

	if path == "/computeMetadata/v1/project/project-id" {
		fmt.Fprint(w, ProjectID)
		return
	}

	if path == "/computeMetadata/v1/project/numeric-project-id" {
		fmt.Fprint(w, NumericProjectID)
		return
	}

	if path == "/computeMetadata/v1/instance/zone" {
		fmt.Fprint(w, fmt.Sprintf("projects/%s/zones/%s-1", NumericProjectID, Region))
		return
	}

	if path == "/computeMetadata/v1/instance/region" {
		fmt.Fprint(w, Region)
		return
	}

	if path == "/computeMetadata/v1/instance/service-accounts/default/token" {
		fmt.Fprint(w, AccessToken)
		return
	}

	if path == "/computeMetadata/v1/instance/service-accounts/default/identity" {
		fmt.Fprint(w, IDToken)
		return
	}
}
