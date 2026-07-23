package dto

import (
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type MessageResponse struct {
	ID              string `json:"id"`
	ConversationID  string `json:"conversation_id"`
	SenderID        string `json:"sender_id"`
	Type            string `json:"type"`
	Text            string `json:"text,omitempty"`
	ClientMessageID string `json:"client_message_id"`
	CreatedAt       string `json:"created_at"`
}

type MessagePageResponse struct {
	Data       []MessageResponse `json:"data"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

func NewMessagePageResponse(messages []entity.Message, hasMore bool, nextCursor string) MessagePageResponse {
	data := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		data = append(data, MessageResponse{ID: message.ID, ConversationID: message.ConversationID,
			SenderID: message.SenderID, Type: message.Type, Text: message.Text,
			ClientMessageID: message.ClientMessageID, CreatedAt: message.CreatedAt.Format(time.RFC3339Nano)})
	}
	return MessagePageResponse{Data: data, NextCursor: nextCursor, HasMore: hasMore}
}
