/*
Copyright © 2018-2021 Neil Hemming
*/

package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReplyServiceUnavailableReturns500(t *testing.T) {
	w := httptest.NewRecorder()

	replyServiceUnavailable(w)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status code %d expected, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestReplyNotFoundReturns404(t *testing.T) {
	w := httptest.NewRecorder()

	replyNotFound(w)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status code %d expected, got %d", http.StatusNotFound, w.Code)
	}
}

func TestReplyInvalidReturns404(t *testing.T) {
	w := httptest.NewRecorder()

	replyInvalid(w)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status code %d expected, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestReplyWithErrorIsJSON(t *testing.T) {
	w := httptest.NewRecorder()

	err := replyWithError(w, 999, "quick")

	if err != nil {
		t.Error("Encoding error")
	}

	if w.Code != 999 {
		t.Errorf("Status code %d expected, got %d", 999, w.Code)
	}

	data := make(map[string]interface{})

	err = json.NewDecoder(w.Body).Decode(&data)
	if err != nil {
		t.Error("Decoding error")
	}

	if s, ok := data["error"].(string); !ok || s != "quick" {
		t.Error("Decoding error field", s, ok)
	}
	if s, ok := data["error_description"].(string); !ok || s != "quick" {
		t.Error("Decoding error_description field", s, ok)
	}
	// not type shift to JSON type
	if v, ok := data["error_code"].(float64); !ok || v != 999 {
		t.Error("Decoding error_code field", v, ok)
	}
}
