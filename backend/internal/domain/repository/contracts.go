package repository

import (
	"context"
	"time"

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

type MediaUploadRepository interface {
	Create(context.Context, *entity.MediaUpload) error
	FindForUser(context.Context, string, string) (*entity.MediaUpload, error)
}

type UploadInspector interface {
	VerifyObject(context.Context, string, string, int64) error
}

type DownloadSigner interface {
	PresignGet(context.Context, string, time.Duration) (string, error)
}

type MediaAttachmentRepository interface {
	FindAttachmentForMember(context.Context, string, string) (*entity.MessageAttachment, error)
}

type UploadSigner interface {
	PresignPut(context.Context, string, string, int64, time.Duration) (string, map[string]string, error)
}

type ConversationRepository interface {
	ListForUser(context.Context, string, int) ([]entity.ConversationSummary, error)
	CreateDirect(context.Context, string, string, string) (*entity.ConversationSummary, bool, error)
	CreateGroup(context.Context, string, entity.CreateGroupInput) (*entity.ConversationSummary, error)
	AddMembers(context.Context, string, string, []string) ([]string, error)
	RemoveMember(context.Context, string, string, string) error
}

type MessageRepository interface {
	ListForMember(context.Context, string, string, *time.Time, string, int) ([]entity.Message, error)
	ListAfterForMember(context.Context, string, string, time.Time, string, int) ([]entity.Message, error)
	Send(context.Context, entity.Message) (*entity.SendMessageResult, error)
	RecipientIDsForMember(context.Context, string, string) ([]string, error)
	PeerIDsForUser(context.Context, string) ([]string, error)
	MarkRead(context.Context, string, string, string) error
}

type MessagePublisher interface {
	PublishMessage(context.Context, entity.MessageEvent) error
}

type RealtimePublisher interface {
	PublishRealtime(context.Context, entity.RealtimeEvent) error
}

type PresenceTracker interface {
	ConnectPresence(context.Context, string, string) (bool, error)
	TouchPresence(context.Context, string, string) error
	DisconnectPresence(context.Context, string, string) (bool, error)
	OnlinePresence(context.Context, []string) (map[string]bool, error)
}
