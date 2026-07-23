package repository

import (
	"context"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
)

type Message struct{ db *gorm.DB }

func NewMessage(db *gorm.DB) *Message { return &Message{db: db} }

func (r *Message) ListForMember(ctx context.Context, conversationID, userID string, before *time.Time, beforeID string, limit int) ([]entity.Message, error) {
	var memberCount int64
	err := r.db.WithContext(ctx).Model(&gormmodel.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).Count(&memberCount).Error
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, domainerrors.ErrForbidden
	}

	query := r.db.WithContext(ctx).Model(&gormmodel.Message{}).Where("conversation_id = ?", conversationID)
	if before != nil {
		query = query.Where("created_at < ? OR (created_at = ? AND id < ?)", *before, *before, beforeID)
	}
	var models []gormmodel.Message
	if err := query.Order("created_at DESC, id DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	messages := make([]entity.Message, 0, len(models))
	for _, model := range models {
		text := ""
		if model.Text != nil {
			text = *model.Text
		}
		messages = append(messages, entity.Message{ID: model.ID, ConversationID: model.ConversationID,
			SenderID: model.SenderID, Type: model.Type, Text: text,
			ClientMessageID: model.ClientMessageID, CreatedAt: model.CreatedAt})
	}
	return messages, nil
}
