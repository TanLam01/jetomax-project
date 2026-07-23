package dto

import (
	"time"

	authusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/auth"
)

type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarKey   string `json:"avatar_key"`
}

type SessionResponse struct {
	User             UserResponse `json:"user"`
	AccessToken      string       `json:"access_token"`
	RefreshToken     string       `json:"refresh_token"`
	TokenType        string       `json:"token_type"`
	AccessExpiresAt  string       `json:"access_expires_at"`
	RefreshExpiresAt string       `json:"refresh_expires_at"`
}

func NewSessionResponse(session *authusecase.Session) SessionResponse {
	return SessionResponse{
		User: UserResponse{
			ID:          session.User.ID,
			Email:       session.User.Email,
			DisplayName: session.User.DisplayName,
			AvatarKey:   session.User.AvatarKey,
		},
		AccessToken:      session.AccessToken,
		RefreshToken:     session.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresAt:  session.AccessExpiresAt.Format(time.RFC3339),
		RefreshExpiresAt: session.RefreshExpiresAt.Format(time.RFC3339),
	}
}
