package websocket

import (
	"encoding/json"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type ClientEvent struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload"`
}

type ServerEvent struct {
	Type      string `json:"type"`
	EventID   string `json:"event_id"`
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
	Payload   any    `json:"payload"`
}

type SendMessagePayload struct {
	ConversationID  string `json:"conversation_id"`
	ClientMessageID string `json:"client_message_id"`
	MessageType     string `json:"message_type"`
	Text            string `json:"text"`
	UploadID        string `json:"upload_id"`
}

type ConversationPayload struct {
	ConversationID string `json:"conversation_id"`
}

type ConversationReadPayload struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
}

type MessagePayload struct {
	ID              string             `json:"id"`
	ConversationID  string             `json:"conversation_id"`
	SenderID        string             `json:"sender_id"`
	Type            string             `json:"type"`
	Text            string             `json:"text,omitempty"`
	ClientMessageID string             `json:"client_message_id"`
	CreatedAt       string             `json:"created_at"`
	Attachment      *AttachmentPayload `json:"attachment,omitempty"`
}

type AttachmentPayload struct {
	ID        string `json:"id"`
	ObjectKey string `json:"object_key"`
	MIMEType  string `json:"mime_type"`
	Size      int64  `json:"size"`
}

func messagePayload(message entity.Message) MessagePayload {
	payload := MessagePayload{ID: message.ID, ConversationID: message.ConversationID, SenderID: message.SenderID,
		Type: message.Type, Text: message.Text, ClientMessageID: message.ClientMessageID,
		CreatedAt: message.CreatedAt.Format(time.RFC3339Nano)}
	if message.Attachment != nil {
		payload.Attachment = &AttachmentPayload{ID: message.Attachment.ID, ObjectKey: message.Attachment.ObjectKey,
			MIMEType: message.Attachment.MIMEType, Size: message.Attachment.Size}
	}
	return payload
}
