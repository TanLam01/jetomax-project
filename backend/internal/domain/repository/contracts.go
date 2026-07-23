package repository

import (
	"context"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type UserRepository interface {
	FindByID(context.Context, string) (*entity.User, error)
	Search(context.Context, string, string, int) ([]entity.User, error)
}

type AuthRepository interface {
	CreateUserWithRefreshToken(context.Context, *entity.User, *entity.RefreshToken) error
	FindUserByEmail(context.Context, string) (*entity.User, error)
	FindUserByRefreshTokenHash(context.Context, string) (*entity.User, *entity.RefreshToken, error)
	CreateRefreshToken(context.Context, *entity.RefreshToken) error
	RotateRefreshToken(context.Context, string, *entity.RefreshToken) error
	RevokeRefreshToken(context.Context, string) error
}

type ErrorRecorder interface {
	Record(context.Context, entity.RequestError) error
}

type ConversationRepository interface {
	ListForUser(context.Context, string, int) ([]entity.ConversationSummary, error)
	CreateDirect(context.Context, string, string, string) (*entity.ConversationSummary, bool, error)
	CreateGroup(context.Context, string, entity.CreateGroupInput) (*entity.ConversationSummary, error)
}

type MessageRepository interface {
	Create(context.Context, *entity.Message) error
	ListByConversation(context.Context, string, string, int) ([]entity.Message, error)
}
