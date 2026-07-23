package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
)

const (
	errorCodeKey    = "safe_error_code"
	errorMessageKey = "safe_error_message"
)

func ErrorAudit(recorder repository.ErrorRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := requestID(c.GetHeader("X-Request-ID"))
		c.Header("X-Request-ID", requestID)
		c.Next()
		if c.Writer.Status() < http.StatusBadRequest {
			return
		}

		code, _ := c.Get(errorCodeKey)
		message, _ := c.Get(errorMessageKey)
		if code == nil {
			code = "http_error"
		}
		if message == nil {
			message = http.StatusText(c.Writer.Status())
		}
		event := entity.RequestError{
			RequestID: requestID, Method: c.Request.Method, Path: c.Request.URL.Path,
			Status: c.Writer.Status(), Code: code.(string), Message: message.(string),
			ClientIP: c.ClientIP(), UserAgent: truncate(c.Request.UserAgent(), 512), CreatedAt: time.Now().UTC(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := recorder.Record(ctx, event); err != nil {
			slog.Error("record failed request", "request_id", requestID, "error", err)
		}
	}
}

func setSafeError(c *gin.Context, code, message string) {
	c.Set(errorCodeKey, code)
	c.Set(errorMessageKey, message)
}

func requestID(candidate string) string {
	if candidate != "" && len(candidate) <= 128 && !strings.ContainsAny(candidate, "\r\n") {
		return candidate
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw)
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
