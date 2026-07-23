package entity

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	AvatarKey    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Conversation struct {
	ID        string
	Type      string
	Name      string
	AvatarKey string
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateGroupInput struct {
	Name      string
	AvatarKey string
	MemberIDs []string
}

type ConversationSummary struct {
	Conversation Conversation
	Role         string
	UnreadCount  int64
	LastMessage  *Message
}

type Message struct {
	ID              string
	ConversationID  string
	SenderID        string
	Type            string
	Text            string
	ClientMessageID string
	CreatedAt       time.Time
}
