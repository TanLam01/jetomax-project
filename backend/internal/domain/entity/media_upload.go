package entity

import "time"

type MediaUpload struct {
	ID           string
	UserID       string
	ObjectKey    string
	OriginalName string
	MIMEType     string
	Size         int64
	Status       string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
