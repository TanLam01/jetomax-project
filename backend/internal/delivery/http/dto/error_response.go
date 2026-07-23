package dto

import "net/http"

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{Error: ErrorDetail{Code: code, Message: message}}
}

// NewPublicErrorResponse returns a stable, non-sensitive error for API clients.
// Detailed causes are retained only by the server-side error audit.
func NewPublicErrorResponse(status int) ErrorResponse {
	code, message := "internal_error", "Internal server error"
	switch status {
	case http.StatusBadRequest:
		code, message = "bad_request", "Bad request"
	case http.StatusUnauthorized:
		code, message = "unauthorized", "Unauthorized"
	case http.StatusForbidden:
		code, message = "forbidden", "Forbidden"
	case http.StatusNotFound:
		code, message = "not_found", "Not found"
	case http.StatusConflict:
		code, message = "conflict", "Conflict"
	case http.StatusTooManyRequests:
		code, message = "too_many_requests", "Too many requests"
	case http.StatusServiceUnavailable:
		code, message = "service_unavailable", "Service unavailable"
	}
	return NewErrorResponse(code, message)
}
