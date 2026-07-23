package entity

import "time"

// RequestError is the framework-independent audit record for a failed HTTP request.
// It intentionally excludes request and response bodies to avoid storing credentials.
type RequestError struct {
	ID        string    `json:"id,omitempty"`
	RequestID string    `json:"request_id"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	Code      string    `json:"error_code"`
	Message   string    `json:"message"`
	ClientIP  string    `json:"client_ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}
