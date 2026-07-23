package dto

import "github.com/jetomax/realtime-chat/backend/internal/domain/entity"

type UserListResponse struct {
	Data []UserResponse `json:"data"`
}

func NewUserResponse(user entity.User) UserResponse {
	return UserResponse{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, AvatarKey: user.AvatarKey}
}

func NewUserListResponse(users []entity.User) UserListResponse {
	data := make([]UserResponse, 0, len(users))
	for _, user := range users {
		data = append(data, NewUserResponse(user))
	}
	return UserListResponse{Data: data}
}
