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
	PeerUserID   string
}

type ConversationMember struct {
	UserID      string
	Email       string
	DisplayName string
	AvatarKey   string
	Role        string
	JoinedAt    time.Time
}

type Message struct {
	ID              string
	ConversationID  string
	SenderID        string
	Type            string
	Text            string
	ClientMessageID string
	CreatedAt       time.Time
	Attachment      *MessageAttachment
}

type MessageAttachment struct {
	ID        string
	MessageID string
	UploadID  string
	ObjectKey string
	MIMEType  string
	Size      int64
}

type SendMessageResult struct {
	Message      Message
	RecipientIDs []string
	Created      bool
}

type MessageEvent struct {
	EventID      string    `json:"event_id"`
	RecipientIDs []string  `json:"recipient_user_ids"`
	Message      Message   `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}

type RealtimeEvent struct {
	Type         string    `json:"type"`
	EventID      string    `json:"event_id"`
	RequestID    string    `json:"request_id"`
	RecipientIDs []string  `json:"recipient_user_ids"`
	Timestamp    time.Time `json:"timestamp"`
	Payload      any       `json:"payload"`
}
