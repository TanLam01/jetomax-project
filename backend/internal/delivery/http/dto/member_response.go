package dto

import (
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type AddMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required"`
}

type AddMembersResponse struct {
	AddedUserIDs []string `json:"added_user_ids"`
	AddedCount   int      `json:"added_count"`
}

func NewAddMembersResponse(userIDs []string) AddMembersResponse {
	if userIDs == nil {
		userIDs = []string{}
	}
	return AddMembersResponse{AddedUserIDs: userIDs, AddedCount: len(userIDs)}
}

type ConversationMemberResponse struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarKey   string `json:"avatar_key,omitempty"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
}

type ConversationMemberListResponse struct {
	Data []ConversationMemberResponse `json:"data"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type TransferOwnershipRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

func NewConversationMemberListResponse(members []entity.ConversationMember) ConversationMemberListResponse {
	data := make([]ConversationMemberResponse, 0, len(members))
	for _, member := range members {
		data = append(data, ConversationMemberResponse{UserID: member.UserID, Email: member.Email,
			DisplayName: member.DisplayName, AvatarKey: member.AvatarKey, Role: member.Role,
			JoinedAt: member.JoinedAt.Format(time.RFC3339)})
	}
	return ConversationMemberListResponse{Data: data}
}
