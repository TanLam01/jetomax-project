package gormmodel

import "time"

type Conversation struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Type      string `gorm:"type:conversation_type;not null"`
	Name      *string
	AvatarKey *string
	DirectKey *string `gorm:"uniqueIndex"`
	CreatedBy string  `gorm:"type:uuid;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Conversation) TableName() string { return "conversations" }

type ConversationMember struct {
	ConversationID    string `gorm:"type:uuid;primaryKey"`
	UserID            string `gorm:"type:uuid;primaryKey"`
	Role              string `gorm:"type:member_role;not null"`
	JoinedAt          time.Time
	LastReadMessageID *string `gorm:"type:uuid"`
}

func (ConversationMember) TableName() string { return "conversation_members" }
