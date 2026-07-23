package dto

import (
	"net/http"
	"testing"
)

func TestNewPublicErrorResponse(t *testing.T) {
	tests := []struct {
		status  int
		code    string
		message string
	}{
		{http.StatusBadRequest, "bad_request", "Bad request"},
		{http.StatusUnauthorized, "unauthorized", "Unauthorized"},
		{http.StatusNotFound, "not_found", "Not found"},
		{http.StatusConflict, "conflict", "Conflict"},
		{http.StatusInternalServerError, "internal_error", "Internal server error"},
		{http.StatusServiceUnavailable, "service_unavailable", "Service unavailable"},
	}
	for _, test := range tests {
		response := NewPublicErrorResponse(test.status)
		if response.Error.Code != test.code || response.Error.Message != test.message {
			t.Errorf("status %d: got %#v", test.status, response.Error)
		}
	}
}
