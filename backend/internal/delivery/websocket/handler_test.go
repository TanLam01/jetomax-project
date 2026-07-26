package websocket

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
)

func TestHandlerKeepsConnectionOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(NewHub(), nil)
	router.GET("/ws", func(c *gin.Context) {
		c.Set("authenticated_user_id", "user-1")
		handler.Connect(c)
	})

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dialCancel()
	conn, response, err := websocket.Dial(dialCtx, "ws"+server.URL[len("http"):]+"/ws", nil)
	if err != nil {
		t.Fatalf("dial websocket: %v (response=%v)", err, response)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request := ClientEvent{Type: "unknown", RequestID: "request-1", Payload: json.RawMessage(`{}`)}
	if err := wsjson.Write(ctx, conn, request); err != nil {
		t.Fatalf("write websocket event: %v", err)
	}
	var responseEvent ServerEvent
	if err := wsjson.Read(ctx, conn, &responseEvent); err != nil {
		t.Fatalf("read websocket event: %v", err)
	}
	if responseEvent.Type != "error" || responseEvent.EventID == "" || responseEvent.RequestID != request.RequestID {
		t.Fatalf("unexpected response event: %+v", responseEvent)
	}
}
