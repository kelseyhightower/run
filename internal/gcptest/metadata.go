package gcptest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	ID               = "00bf4bf02db3546595153c211fec26b688516bf1b609f7d5e2f5d9ae17d1bcbaf6ce6c0c1b1168abf1ab255125e84e085336a36ae5715b0f95e7"
	NumericProjectID = "123456789"
	ProjectID        = "test"
	Region           = "test"
)

var AccessToken = AccessTokenResponse{
	AccessToken: "ya29.AHES6ZRVmB7fkLtd1XTmq6mo0S1wqZZi3-Lh_s-6Uw7p8vtgSwg",
	ExpiresIn:   3484,
	TokenType:   "Bearer",
}

// AccessTokenResponse holds a GCP access token.
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

const (
	IDToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.Et9HFtf9R3GEMA0IICOfFMVXY7kkTX1wr4qCyhIf58U"
)

func BrokenMetadataHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/computeMetadata/v1/instance/service-accounts/default/token" {
		w.Write([]byte("{\"broken\""))
		return
	}

	http.Error(w, "", 500)
	return
}

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
		data, err := json.Marshal(AccessToken)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(data)
		return
	}

	if path == "/computeMetadata/v1/instance/service-accounts/default/identity" {
		fmt.Fprint(w, IDToken)
		return
	}
}
