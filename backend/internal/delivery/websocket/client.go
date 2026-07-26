package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
	messageusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/message"
)

type Client struct {
	userID       string
	conn         *websocket.Conn
	hub          *Hub
	service      *messageusecase.Service
	send         chan ServerEvent
	ctx          context.Context
	cancel       context.CancelFunc
	errors       repository.ErrorRecorder
	request      RequestMetadata
	connectionID string
}

type RequestMetadata struct {
	Path, ClientIP, UserAgent string
}

func NewClient(parent context.Context, userID string, conn *websocket.Conn, hub *Hub, service *messageusecase.Service, errorRecorder repository.ErrorRecorder, request RequestMetadata) *Client {
	ctx, cancel := context.WithCancel(parent)
	return &Client{userID: userID, conn: conn, hub: hub, service: service,
		send: make(chan ServerEvent, 64), ctx: ctx, cancel: cancel, errors: errorRecorder, request: request,
		connectionID: uuid.NewString()}
}

func (c *Client) Run() {
	c.hub.Register(c)
	if c.service != nil {
		presenceCtx, presenceCancel := context.WithTimeout(c.ctx, 2*time.Second)
		if err := c.service.PresenceConnected(presenceCtx, c.userID, c.connectionID); err != nil {
			slog.Error("connect websocket presence", "user_id", c.userID, "error", err)
		}
		if onlinePeers, err := c.service.OnlinePeers(presenceCtx, c.userID); err != nil {
			slog.Error("load websocket presence snapshot", "user_id", c.userID, "error", err)
		} else {
			for _, peerID := range onlinePeers {
				c.Enqueue(ServerEvent{Type: "presence.changed", EventID: uuid.NewString(), RequestID: "",
					Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Payload: map[string]any{"user_id": peerID, "online": true}})
			}
		}
		presenceCancel()
	}
	writeDone := make(chan struct{})
	go func() { defer close(writeDone); c.writeLoop() }()
	c.readLoop()
	c.cancel()
	c.hub.Unregister(c)
	if c.service != nil {
		presenceCtx, presenceCancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := c.service.PresenceDisconnected(presenceCtx, c.userID, c.connectionID); err != nil {
			slog.Error("disconnect websocket presence", "user_id", c.userID, "error", err)
		}
		presenceCancel()
	}
	<-writeDone
	_ = c.conn.Close(websocket.StatusNormalClosure, "connection closed")
}

func (c *Client) Enqueue(event ServerEvent) {
	select {
	case c.send <- event:
	default:
		c.cancel()
	}
}

func (c *Client) readLoop() {
	c.conn.SetReadLimit(64 * 1024)
	for {
		var event ClientEvent
		if err := wsjson.Read(c.ctx, c.conn, &event); err != nil {
			status := websocket.CloseStatus(err)
			if status != websocket.StatusNormalClosure && status != websocket.StatusGoingAway && !errors.Is(err, context.Canceled) {
				slog.Warn("websocket read loop ended", "user_id", c.userID, "error", err)
			}
			return
		}
		c.handle(event)
	}
}

func (c *Client) handle(event ClientEvent) {
	if event.RequestID == "" || len(event.RequestID) > 128 {
		c.sendFailure(event.RequestID, http.StatusBadRequest, "bad_request", errors.New("missing or invalid request_id"))
		return
	}
	switch event.Type {
	case "message.send":
		var payload SendMessagePayload
		if json.Unmarshal(event.Payload, &payload) != nil {
			c.sendFailure(event.RequestID, http.StatusBadRequest, "bad_request", errors.New("invalid message.send payload"))
			return
		}
		message, err := c.service.Send(c.ctx, c.userID, messageusecase.SendInput{
			ConversationID: payload.ConversationID, ClientMessageID: payload.ClientMessageID,
			Type: payload.MessageType, Text: payload.Text, UploadID: payload.UploadID,
		})
		if err != nil {
			slog.Warn("websocket message send failed", "user_id", c.userID, "request_id", event.RequestID, "error", err)
			status, code := http.StatusInternalServerError, "internal_error"
			switch {
			case errors.Is(err, domainerrors.ErrValidation):
				status, code = http.StatusBadRequest, "bad_request"
			case errors.Is(err, domainerrors.ErrForbidden):
				status, code = http.StatusForbidden, "forbidden"
			case errors.Is(err, domainerrors.ErrNotFound):
				status, code = http.StatusNotFound, "not_found"
			case errors.Is(err, domainerrors.ErrConflict):
				status, code = http.StatusConflict, "conflict"
			}
			c.sendFailure(event.RequestID, status, code, err)
			return
		}
		c.Enqueue(ServerEvent{Type: "message.ack", EventID: uuid.NewString(), RequestID: event.RequestID,
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Payload: map[string]any{
				"message_id": message.ID, "client_message_id": message.ClientMessageID, "status": "sent",
			}})
	case "typing.start", "typing.stop":
		var payload ConversationPayload
		if json.Unmarshal(event.Payload, &payload) != nil {
			c.sendFailure(event.RequestID, http.StatusBadRequest, "bad_request", errors.New("invalid typing payload"))
			return
		}
		if err := c.service.PublishTyping(c.ctx, c.userID, payload.ConversationID, event.RequestID, event.Type == "typing.start"); err != nil {
			c.handleServiceFailure(event.RequestID, err)
		}
	case "conversation.read":
		var payload ConversationReadPayload
		if json.Unmarshal(event.Payload, &payload) != nil {
			c.sendFailure(event.RequestID, http.StatusBadRequest, "bad_request", errors.New("invalid conversation.read payload"))
			return
		}
		if err := c.service.MarkRead(c.ctx, c.userID, payload.ConversationID, payload.MessageID, event.RequestID); err != nil {
			c.handleServiceFailure(event.RequestID, err)
		}
	default:
		c.sendFailure(event.RequestID, http.StatusBadRequest, "bad_request", errors.New("unsupported websocket event type"))
	}
}

func (c *Client) handleServiceFailure(requestID string, err error) {
	slog.Warn("websocket event failed", "user_id", c.userID, "request_id", requestID, "error", err)
	status, code := http.StatusInternalServerError, "internal_error"
	switch {
	case errors.Is(err, domainerrors.ErrValidation):
		status, code = http.StatusBadRequest, "bad_request"
	case errors.Is(err, domainerrors.ErrForbidden):
		status, code = http.StatusForbidden, "forbidden"
	case errors.Is(err, domainerrors.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, domainerrors.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	}
	c.sendFailure(requestID, status, code, err)
}

func (c *Client) sendFailure(requestID string, status int, code string, cause error) {
	if requestID == "" || len(requestID) > 128 {
		requestID = uuid.NewString()
	}
	message := http.StatusText(status)
	c.Enqueue(ServerEvent{Type: "error", EventID: uuid.NewString(), RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano), Payload: map[string]string{"code": code, "message": message}})
	if c.errors == nil {
		return
	}
	event := entity.RequestError{RequestID: requestID, Method: "WEBSOCKET", Path: c.request.Path,
		Status: status, Code: code, Message: cause.Error(), ClientIP: c.request.ClientIP,
		UserAgent: truncate(c.request.UserAgent, 512), CreatedAt: time.Now().UTC()}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.errors.Record(ctx, event); err != nil {
		slog.Error("record failed websocket event", "request_id", requestID, "error", err)
	}
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func (c *Client) writeLoop() {
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case event := <-c.send:
			ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
			err := wsjson.Write(ctx, c.conn, event)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		case <-heartbeat.C:
			ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
			presenceCtx, presenceCancel := context.WithTimeout(c.ctx, 2*time.Second)
			if c.service != nil {
				if err := c.service.TouchPresence(presenceCtx, c.userID, c.connectionID); err != nil {
					slog.Error("refresh websocket presence", "user_id", c.userID, "error", err)
				}
			}
			presenceCancel()
		}
	}
}
