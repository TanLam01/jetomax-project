package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadiness(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	NewRouter("test", nil, nil, nil, nil, nil, nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}
