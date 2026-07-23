package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	authusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/auth"
)

func TestNewSessionResponseDoesNotExposePasswordHash(t *testing.T) {
	session := &authusecase.Session{
		User: entity.User{
			ID:           "user-id",
			Email:        "user@example.com",
			DisplayName:  "User",
			PasswordHash: "must-not-be-serialized",
		},
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		AccessExpiresAt:  time.Now(),
		RefreshExpiresAt: time.Now(),
	}

	encoded, err := json.Marshal(NewSessionResponse(session))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if strings.Contains(string(encoded), "must-not-be-serialized") || strings.Contains(string(encoded), "password") {
		t.Fatalf("session response exposed password data: %s", encoded)
	}
}
