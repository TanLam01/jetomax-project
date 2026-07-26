package repository

import (
	"context"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
)

type MediaUpload struct{ db *gorm.DB }

func NewMediaUpload(db *gorm.DB) *MediaUpload { return &MediaUpload{db: db} }

func (r *MediaUpload) Create(ctx context.Context, upload *entity.MediaUpload) error {
	model := gormmodel.MediaUpload{ID: upload.ID, UserID: upload.UserID, ObjectKey: upload.ObjectKey,
		OriginalName: upload.OriginalName, MIMEType: upload.MIMEType, Size: upload.Size,
		Status: upload.Status, ExpiresAt: upload.ExpiresAt}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return translate(err)
	}
	upload.CreatedAt = model.CreatedAt
	return nil
}

func (r *MediaUpload) FindForUser(ctx context.Context, uploadID, userID string) (*entity.MediaUpload, error) {
	var model gormmodel.MediaUpload
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", uploadID, userID).First(&model).Error; err != nil {
		return nil, translate(err)
	}
	return &entity.MediaUpload{ID: model.ID, UserID: model.UserID, ObjectKey: model.ObjectKey,
		OriginalName: model.OriginalName, MIMEType: model.MIMEType, Size: model.Size,
		Status: model.Status, ExpiresAt: model.ExpiresAt, CreatedAt: model.CreatedAt}, nil
}

func (r *MediaUpload) FindAttachmentForMember(ctx context.Context, attachmentID, userID string) (*entity.MessageAttachment, error) {
	var model gormmodel.MessageAttachment
	err := r.db.WithContext(ctx).Model(&gormmodel.MessageAttachment{}).
		Joins("JOIN messages ON messages.id = message_attachments.message_id").
		Joins("JOIN conversation_members ON conversation_members.conversation_id = messages.conversation_id AND conversation_members.user_id = ?", userID).
		Where("message_attachments.id = ?", attachmentID).First(&model).Error
	if err != nil {
		return nil, translate(err)
	}
	return &entity.MessageAttachment{ID: model.ID, MessageID: model.MessageID, UploadID: model.UploadID,
		ObjectKey: model.ObjectKey, MIMEType: model.MIMEType, Size: model.Size}, nil
}
