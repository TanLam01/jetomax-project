package gormmodel

type MessageAttachment struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MessageID string `gorm:"type:uuid;not null;uniqueIndex"`
	UploadID  string `gorm:"type:uuid;not null;uniqueIndex"`
	ObjectKey string `gorm:"not null"`
	MIMEType  string `gorm:"not null"`
	Size      int64  `gorm:"not null"`
}

func (MessageAttachment) TableName() string { return "message_attachments" }
