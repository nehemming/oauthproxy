/*
Copyright © 2018-2021 Neil Hemming
*/

package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewRuntime(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")
	rt := newRuntime(context.Background(), settings)
	defer rt.close()

	if rt.isStopping {
		t.Error("Stopping set")
	}

	// Validate rt vs settings
	if rt.ttl != settings.CacheTTL {
		t.Errorf("Mismatch ttl %d vs CacheTTL %d", rt.ttl, settings.CacheTTL)
	}
	if rt.requestTimeout != settings.RequestTimeout {
		t.Errorf("Mismatch requestTimeout %d vs RequestTimeout %d", rt.requestTimeout, settings.RequestTimeout)
	}
	if rt.endpoint != settings.Endpoint {
		t.Errorf("Mismatch endpoint %s vs Endpoint %s", rt.endpoint, settings.Endpoint)
	}
	if rt.houseKeeperPeriod != settings.CacheTTL {
		t.Errorf("Mismatch houseKeeperPeriod %d vs CacheTTL %d", rt.houseKeeperPeriod, settings.CacheTTL)
	}
}

func TestCriticalError(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")
	rt := newRuntime(context.Background(), settings)

	defer rt.close()

	expected := errors.New("test")

	rt.criticalError(expected)

	<-rt.done()

	if rt.err != expected {
		t.Error("No critical error", rt.err)
	}
}

func TestLoggingInfo(t *testing.T) {

	called := false

	fn := func(isErr bool, fmt string, args ...interface{}) {
		if called {
			return /// stop additional logging
		}
		if isErr {
			t.Error("Not Info")
		}
		if fmt != "test" {
			t.Error("Fmt err", fmt)
		}
		called = true
	}

	settings := DefaultSettings().WithEndpoint("test").WithLogger(fn)
	rt := newRuntime(context.Background(), settings)
	defer rt.close()

	rt.logInfo("test")

	if !called {
		t.Error("Not called")
	}
}

func TestLoggingError(t *testing.T) {

	called := false

	fn := func(isErr bool, fmt string, args ...interface{}) {
		if called {
			return /// stop additional logging
		}
		if !isErr {
			t.Error("Not Error")
		}
		if fmt != "err" {
			t.Error("Fmt err", fmt)
		}
		called = true
	}

	settings := DefaultSettings().WithEndpoint("test").WithLogger(fn)
	rt := newRuntime(context.Background(), settings)
	defer rt.close()

	rt.logError("err")

	if !called {
		t.Error("Not called")
	}
}

func TestParseRequestNoMatch(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")
	rt := newRuntime(context.Background(), settings)
	defer rt.close()
	reader := strings.NewReader("client_id=123&client_secret=456&grant_type=password&password=abc&username=def")
	req, _ := http.NewRequest("POST", "http:/something", reader)

	w := httptest.NewRecorder()
	tr, match := rt.parseRequest(w, req)

	if match {
		t.Error("Not meant to match")
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected code %d got %d", http.StatusNotFound, w.Code)
	}

	if tr != (tokenRequest{path: "/something"}) {
		t.Error("Unexpected token returned", tr)
	}
}

func TestParseRequestMatchBadRequestFail(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")
	rt := newRuntime(context.Background(), settings)
	defer rt.close()
	reader := strings.NewReader("client_id=123&client_secret=456&grant_type=password&password=p1&scope=alpha+bravo&username=u1")
	req, _ := http.NewRequest("POST", "http:/something/token", reader)
	//no header coontent type

	w := httptest.NewRecorder()
	_, match := rt.parseRequest(w, req)

	if match {
		t.Error("Unexpected a match")
	}

	if w.Code != http.StatusBadRequest {
		t.Error("Status not bad", w.Code)
	}

}

func TestParseRequestMatchBody(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")
	rt := newRuntime(context.Background(), settings)
	defer rt.close()
	reader := strings.NewReader("client_id=123&client_secret=456&grant_type=password&password=p1&scope=alpha+bravo&username=u1")
	req, _ := http.NewRequest("POST", "http:/something/token", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	tr, match := rt.parseRequest(w, req)

	if !match {
		t.Error("Expected a match")
	}

	if w.Code != http.StatusOK {
		t.Error("Status set", w.Code)
	}

	expected := tokenRequest{
		path:         "/something/token",
		clientID:     "123",
		clientSecret: "456",
		username:     "u1",
		password:     "p1",
		scopes:       "alpha bravo",
		authMode:     authInBody,
	}

	if tr != expected {
		t.Error("Unexpected token returned", tr)
	}
}

func TestParseRequestMatchHeader(t *testing.T) {

	settings := DefaultSettings().WithEndpoint("test")
	rt := newRuntime(context.Background(), settings)
	defer rt.close()
	reader := strings.NewReader("grant_type=password&password=p1&scope=alpha+bravo&username=u1")
	req, _ := http.NewRequest("POST", "http:/something/token", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("123", "456")

	w := httptest.NewRecorder()
	tr, match := rt.parseRequest(w, req)

	if !match {
		t.Error("Expected a match")
	}

	if w.Code != http.StatusOK {
		t.Error("Status set", w.Code)
	}

	expected := tokenRequest{
		path:         "/something/token",
		clientID:     "123",
		clientSecret: "456",
		username:     "u1",
		password:     "p1",
		scopes:       "alpha bravo",
		authMode:     authInHeader,
	}

	if tr != expected {
		t.Error("Unexpected token returned", tr)
	}
}
