package dto

import (
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type MessagePreviewResponse struct {
	ID        string `json:"id"`
	SenderID  string `json:"sender_id"`
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	CreatedAt string `json:"created_at"`
}

type ConversationResponse struct {
	ID          string                  `json:"id"`
	Type        string                  `json:"type"`
	Name        string                  `json:"name,omitempty"`
	AvatarKey   string                  `json:"avatar_key,omitempty"`
	Role        string                  `json:"role"`
	UnreadCount int64                   `json:"unread_count"`
	LastMessage *MessagePreviewResponse `json:"last_message"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

type ConversationListResponse struct {
	Data []ConversationResponse `json:"data"`
}

type CreateDirectRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type CreateGroupRequest struct {
	Name      string   `json:"name" binding:"required"`
	AvatarKey string   `json:"avatar_key"`
	MemberIDs []string `json:"member_ids" binding:"required"`
}

func NewConversationListResponse(items []entity.ConversationSummary) ConversationListResponse {
	data := make([]ConversationResponse, 0, len(items))
	for _, item := range items {
		data = append(data, NewConversationResponse(item))
	}
	return ConversationListResponse{Data: data}
}

func NewConversationResponse(item entity.ConversationSummary) ConversationResponse {
	response := ConversationResponse{ID: item.Conversation.ID, Type: item.Conversation.Type,
		Name: item.Conversation.Name, AvatarKey: item.Conversation.AvatarKey, Role: item.Role,
		UnreadCount: item.UnreadCount, CreatedAt: item.Conversation.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.Conversation.UpdatedAt.Format(time.RFC3339)}
	if item.LastMessage != nil {
		response.LastMessage = &MessagePreviewResponse{ID: item.LastMessage.ID, SenderID: item.LastMessage.SenderID,
			Type: item.LastMessage.Type, Text: item.LastMessage.Text, CreatedAt: item.LastMessage.CreatedAt.Format(time.RFC3339)}
	}
	return response
}
