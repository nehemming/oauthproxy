/*
Copyright © 2018-2021 Neil Hemming
*/

package proxy

import (
	"testing"
)

func TestPrepareRequestBasicSucceeds(t *testing.T) {
	tr := tokenRequest{}

	req, err := tr.prepareRequest("basic")

	if err != nil {
		t.Error("Expected success, got", err)
	}

	auth := req.Header.Get("Authorization")

	authExpected := "Basic Og=="
	if auth != authExpected {
		t.Errorf("Authorization expected %s, got %s", authExpected, auth)
	}

	len := req.ContentLength
	b := make([]byte, len)

	_, err = req.Body.Read(b)
	if err != nil {
		t.Error("Read error", err)
	}

	expected := "grant_type=password&password=&username="

	got := string(b)
	if got != expected {
		t.Errorf("Body expected %s, got %s", expected, got)
	}
}

func TestPrepareRequestScopesucceeds(t *testing.T) {
	tr := tokenRequest{scopes: "one two"}

	req, err := tr.prepareRequest("scopes")

	if err != nil {
		t.Error("Expected success, got", err)
	}

	len := req.ContentLength
	b := make([]byte, len)

	_, err = req.Body.Read(b)
	if err != nil {
		t.Error("Read error", err)
	}

	expected := "grant_type=password&password=&scope=one+two&username="

	got := string(b)
	if got != expected {
		t.Errorf("Body expected %s, got %s", expected, got)
	}
}

func TestPrepareRequestAuthBodySucceeds(t *testing.T) {
	tr := tokenRequest{
		authMode: authInBody,
	}

	req, err := tr.prepareRequest("body")

	if err != nil {
		t.Error("Expected success, got", err)
	}

	len := req.ContentLength
	b := make([]byte, len)

	_, err = req.Body.Read(b)
	if err != nil {
		t.Error("Read error", err)
	}

	expected := "client_id=&client_secret=&grant_type=password&password=&username="

	got := string(b)
	if got != expected {
		t.Errorf("Body expected %s, got %s", expected, got)
	}
}

func TestGenRequestROUNDTRIP(t *testing.T) {
	tr := tokenRequest{
		clientID:     "123",
		clientSecret: "456",
		username:     "u1",
		password:     "p1",
		scopes:       "alpha bravo",
		authMode:     authInBody,
	}

	req, err := tr.prepareRequest("down")
	if err != nil {
		t.Error("prepareRequest", err)
		return
	}

	if err := req.ParseForm(); err != nil {
		t.Error("ParseForm", err)
	}

	if grantType := req.PostFormValue("grant_type"); grantType != "password" {
		t.Error("grant_type missing")
	}
	if clientID := req.PostFormValue("client_id"); clientID != "123" {
		t.Error("client_id missing")
	}
	if clientSecret := req.PostFormValue("client_secret"); clientSecret != "456" {
		t.Error("client_secret missing")
	}
	if username := req.PostFormValue("username"); username != "u1" {
		t.Error("username missing")
	}
	if password := req.PostFormValue("password"); password != "p1" {
		t.Error("password missing")
	}
}
