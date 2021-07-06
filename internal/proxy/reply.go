/*
Copyright © 2018-2021 Neil Hemming
*/

package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/nehemming/cirocket/pkg/loggee"
)

func replyServiceUnavailable(w http.ResponseWriter) {
	err := replyWithError(w, http.StatusServiceUnavailable, "Service unavailable")
	if err != nil {
		loggee.Warn(err.Error())
	}
}

func replyNotFound(w http.ResponseWriter) {
	err := replyWithError(w, http.StatusNotFound, "Not found")
	if err != nil {
		loggee.Warn(err.Error())
	}
}

func replyInvalid(w http.ResponseWriter) {
	err := replyWithError(w, http.StatusBadRequest, "bad request")
	if err != nil {
		loggee.Warn(err.Error())
	}
}

func replyWithError(w http.ResponseWriter, statusCode int, msg string) error {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	data := make(map[string]interface{})
	data["error"] = msg
	data["error_description"] = msg
	data["error_code"] = statusCode

	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
