package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{clients: make(map[string]map[*Client]struct{}), ctx: ctx, cancel: cancel}
}

func (h *Hub) Context() context.Context { return h.ctx }

func (h *Hub) Close() {
	h.cancel()
	h.mu.RLock()
	clients := make([]*Client, 0)
	for _, userClients := range h.clients {
		for client := range userClients {
			clients = append(clients, client)
		}
	}
	h.mu.RUnlock()
	for _, client := range clients {
		client.cancel()
		_ = client.conn.CloseNow()
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*Client]struct{})
	}
	h.clients[client.userID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients[client.userID], client)
	if len(h.clients[client.userID]) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *Hub) SendToUser(userID string, event ServerEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[userID] {
		client.Enqueue(event)
	}
}

func (h *Hub) BroadcastMessage(event entity.MessageEvent) {
	serverEvent := ServerEvent{Type: "message.created", EventID: event.EventID,
		Timestamp: event.Timestamp.UTC().Format(time.RFC3339Nano), Payload: messagePayload(event.Message)}
	for _, userID := range event.RecipientIDs {
		h.SendToUser(userID, serverEvent)
	}
}

func (h *Hub) BroadcastRealtime(event entity.RealtimeEvent) {
	serverEvent := ServerEvent{Type: event.Type, EventID: event.EventID, RequestID: event.RequestID,
		Timestamp: event.Timestamp.UTC().Format(time.RFC3339Nano), Payload: event.Payload}
	for _, userID := range event.RecipientIDs {
		h.SendToUser(userID, serverEvent)
	}
}
