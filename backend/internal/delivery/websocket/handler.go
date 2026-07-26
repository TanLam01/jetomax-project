package websocket

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
	messageusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/message"
)

type Handler struct {
	hub     *Hub
	service *messageusecase.Service
	errors  repository.ErrorRecorder
}

func NewHandler(hub *Hub, service *messageusecase.Service, errorRecorder ...repository.ErrorRecorder) *Handler {
	handler := &Handler{hub: hub, service: service}
	if len(errorRecorder) > 0 {
		handler.errors = errorRecorder[0]
	}
	return handler
}

func (h *Handler) Connect(c *gin.Context) {
	writer := http.ResponseWriter(c.Writer)
	if unwrapper, ok := writer.(interface{ Unwrap() http.ResponseWriter }); ok {
		writer = unwrapper.Unwrap()
	}
	conn, err := websocket.Accept(writer, c.Request, &websocket.AcceptOptions{CompressionMode: websocket.CompressionDisabled})
	if err != nil {
		slog.Error("accept websocket connection", "error", err)
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Error("websocket handler panic", "error", recovered, "stack", string(debug.Stack()))
			_ = conn.Close(websocket.StatusInternalError, "internal error")
		}
	}()
	client := NewClient(h.hub.Context(), c.GetString("authenticated_user_id"), conn, h.hub, h.service, h.errors, RequestMetadata{
		Path: c.Request.URL.Path, ClientIP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
	})
	client.Run()
}
