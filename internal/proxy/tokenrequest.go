/*
Copyright © 2018-2021 Neil Hemming
*/

package proxy

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	authInHeader = authType(iota)
	authInBody
)

type (
	authType int

	tokenRequest struct {
		path         string
		clientID     string
		clientSecret string
		username     string
		password     string
		scopes       string
		authMode     authType
	}
)

func (tr *tokenRequest) prepareRequest(endpointURL string) (*http.Request, error) {
	v := url.Values{
		"grant_type": {"password"},
		"username":   {tr.username},
		"password":   {tr.password},
	}

	// Embed auth in body
	if tr.authMode == authInBody {
		v.Set("client_id", tr.clientID)
		v.Set("client_secret", tr.clientSecret)
	}

	if len(tr.scopes) > 0 {
		v.Set("scope", tr.scopes)
	}

	url := endpointURL + tr.path

	req, err := http.NewRequest("POST", url, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Embed auth in header
	if tr.authMode == authInHeader {
		req.SetBasicAuth(tr.clientID, tr.clientSecret)
	}

	return req, nil
}
