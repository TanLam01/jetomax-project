package dto

import (
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type MessageResponse struct {
	ID              string                     `json:"id"`
	ConversationID  string                     `json:"conversation_id"`
	SenderID        string                     `json:"sender_id"`
	Type            string                     `json:"type"`
	Text            string                     `json:"text,omitempty"`
	ClientMessageID string                     `json:"client_message_id"`
	CreatedAt       string                     `json:"created_at"`
	Attachment      *MessageAttachmentResponse `json:"attachment,omitempty"`
}

type MessageAttachmentResponse struct {
	ID        string `json:"id"`
	ObjectKey string `json:"object_key"`
	MIMEType  string `json:"mime_type"`
	Size      int64  `json:"size"`
}

type MessagePageResponse struct {
	Data       []MessageResponse `json:"data"`
	NextCursor string            `json:"next_cursor,omitempty"`
	SyncCursor string            `json:"sync_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

func NewMessagePageResponse(messages []entity.Message, hasMore bool, nextCursor, syncCursor string) MessagePageResponse {
	data := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		response := MessageResponse{ID: message.ID, ConversationID: message.ConversationID,
			SenderID: message.SenderID, Type: message.Type, Text: message.Text,
			ClientMessageID: message.ClientMessageID, CreatedAt: message.CreatedAt.Format(time.RFC3339Nano)}
		if message.Attachment != nil {
			response.Attachment = &MessageAttachmentResponse{ID: message.Attachment.ID, ObjectKey: message.Attachment.ObjectKey,
				MIMEType: message.Attachment.MIMEType, Size: message.Attachment.Size}
		}
		data = append(data, response)
	}
	return MessagePageResponse{Data: data, NextCursor: nextCursor, SyncCursor: syncCursor, HasMore: hasMore}
}
