package repository

import (
	"context"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Message struct{ db *gorm.DB }

func NewMessage(db *gorm.DB) *Message { return &Message{db: db} }

func (r *Message) RecipientIDsForMember(ctx context.Context, conversationID, userID string) ([]string, error) {
	var memberCount int64
	if err := r.db.WithContext(ctx).Model(&gormmodel.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).Count(&memberCount).Error; err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, domainerrors.ErrForbidden
	}
	var ids []string
	if err := r.db.WithContext(ctx).Model(&gormmodel.ConversationMember{}).
		Where("conversation_id = ?", conversationID).Pluck("user_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *Message) PeerIDsForUser(ctx context.Context, userID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Table("conversation_members AS mine").Distinct("peer.user_id").
		Joins("JOIN conversation_members AS peer ON peer.conversation_id = mine.conversation_id").
		Where("mine.user_id = ? AND peer.user_id <> ?", userID, userID).Pluck("peer.user_id", &ids).Error
	return ids, err
}

func (r *Message) MarkRead(ctx context.Context, conversationID, userID, messageID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var memberCount int64
		if err := tx.Model(&gormmodel.ConversationMember{}).
			Where("conversation_id = ? AND user_id = ?", conversationID, userID).Count(&memberCount).Error; err != nil {
			return err
		}
		if memberCount == 0 {
			return domainerrors.ErrForbidden
		}
		var target gormmodel.Message
		if err := tx.Where("id = ? AND conversation_id = ?", messageID, conversationID).First(&target).Error; err != nil {
			return translate(err)
		}
		result := tx.Exec(`UPDATE conversation_members AS member
			SET last_read_message_id = ?
			WHERE member.conversation_id = ? AND member.user_id = ?
			  AND (member.last_read_message_id IS NULL OR EXISTS (
			      SELECT 1 FROM messages current_message
			      WHERE current_message.id = member.last_read_message_id
			        AND (current_message.created_at < ? OR (current_message.created_at = ? AND current_message.id <= ?))
			  ))`, messageID, conversationID, userID, target.CreatedAt, target.CreatedAt, target.ID)
		return result.Error
	})
}

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
	attachments := make(map[string]entity.MessageAttachment)
	if len(models) > 0 {
		messageIDs := make([]string, 0, len(models))
		for _, model := range models {
			messageIDs = append(messageIDs, model.ID)
		}
		var attachmentModels []gormmodel.MessageAttachment
		if err := r.db.WithContext(ctx).Where("message_id IN ?", messageIDs).Find(&attachmentModels).Error; err != nil {
			return nil, err
		}
		for _, attachment := range attachmentModels {
			attachments[attachment.MessageID] = entity.MessageAttachment{ID: attachment.ID, MessageID: attachment.MessageID,
				UploadID: attachment.UploadID, ObjectKey: attachment.ObjectKey, MIMEType: attachment.MIMEType, Size: attachment.Size}
		}
	}
	for _, model := range models {
		text := ""
		if model.Text != nil {
			text = *model.Text
		}
		message := entity.Message{ID: model.ID, ConversationID: model.ConversationID,
			SenderID: model.SenderID, Type: model.Type, Text: text,
			ClientMessageID: model.ClientMessageID, CreatedAt: model.CreatedAt}
		if attachment, ok := attachments[model.ID]; ok {
			message.Attachment = &attachment
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (r *Message) ListAfterForMember(ctx context.Context, conversationID, userID string, after time.Time, afterID string, limit int) ([]entity.Message, error) {
	var memberCount int64
	err := r.db.WithContext(ctx).Model(&gormmodel.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).Count(&memberCount).Error
	if err != nil {
		return nil, err
	}
	if memberCount == 0 {
		return nil, domainerrors.ErrForbidden
	}
	var models []gormmodel.Message
	err = r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).
		Where("created_at > ? OR (created_at = ? AND id > ?)", after, after, afterID).
		Order("created_at ASC, id ASC").Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}
	attachments := make(map[string]entity.MessageAttachment)
	if len(models) > 0 {
		messageIDs := make([]string, 0, len(models))
		for _, model := range models {
			messageIDs = append(messageIDs, model.ID)
		}
		var attachmentModels []gormmodel.MessageAttachment
		if err := r.db.WithContext(ctx).Where("message_id IN ?", messageIDs).Find(&attachmentModels).Error; err != nil {
			return nil, err
		}
		for _, attachment := range attachmentModels {
			attachments[attachment.MessageID] = entity.MessageAttachment{ID: attachment.ID, MessageID: attachment.MessageID,
				UploadID: attachment.UploadID, ObjectKey: attachment.ObjectKey, MIMEType: attachment.MIMEType, Size: attachment.Size}
		}
	}
	messages := make([]entity.Message, 0, len(models))
	for _, model := range models {
		text := ""
		if model.Text != nil {
			text = *model.Text
		}
		message := entity.Message{ID: model.ID, ConversationID: model.ConversationID, SenderID: model.SenderID,
			Type: model.Type, Text: text, ClientMessageID: model.ClientMessageID, CreatedAt: model.CreatedAt}
		if attachment, ok := attachments[model.ID]; ok {
			message.Attachment = &attachment
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (r *Message) Send(ctx context.Context, message entity.Message) (*entity.SendMessageResult, error) {
	result := &entity.SendMessageResult{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var memberCount int64
		if err := tx.Model(&gormmodel.ConversationMember{}).
			Where("conversation_id = ? AND user_id = ?", message.ConversationID, message.SenderID).
			Count(&memberCount).Error; err != nil {
			return err
		}
		if memberCount == 0 {
			return domainerrors.ErrForbidden
		}

		text := message.Text
		model := gormmodel.Message{ConversationID: message.ConversationID, SenderID: message.SenderID,
			Type: message.Type, Text: &text, ClientMessageID: message.ClientMessageID}
		create := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "sender_id"}, {Name: "client_message_id"}}, DoNothing: true}).Create(&model)
		if create.Error != nil {
			return translate(create.Error)
		}
		result.Created = create.RowsAffected == 1
		if !result.Created {
			if err := tx.Where("sender_id = ? AND client_message_id = ?", message.SenderID, message.ClientMessageID).First(&model).Error; err != nil {
				return translate(err)
			}
			existingText := ""
			if model.Text != nil {
				existingText = *model.Text
			}
			if model.ConversationID != message.ConversationID || model.Type != message.Type || existingText != message.Text {
				return domainerrors.ErrConflict
			}
			if err := compareStoredAttachment(tx, model.ID, message.Attachment); err != nil {
				return err
			}
		} else {
			if message.Attachment != nil {
				claim := tx.Model(&gormmodel.MediaUpload{}).
					Where("id = ? AND user_id = ? AND status IN ?", message.Attachment.UploadID, message.SenderID, []string{"pending", "uploaded"}).
					Update("status", "uploaded")
				if claim.Error != nil {
					return translate(claim.Error)
				}
				if claim.RowsAffected != 1 {
					return domainerrors.ErrConflict
				}
				attachmentModel := gormmodel.MessageAttachment{MessageID: model.ID, UploadID: message.Attachment.UploadID,
					ObjectKey: message.Attachment.ObjectKey, MIMEType: message.Attachment.MIMEType, Size: message.Attachment.Size}
				if err := tx.Create(&attachmentModel).Error; err != nil {
					return translate(err)
				}
				message.Attachment.ID = attachmentModel.ID
				message.Attachment.MessageID = model.ID
			}
			if err := tx.Model(&gormmodel.Conversation{}).Where("id = ?", message.ConversationID).
				Update("updated_at", model.CreatedAt).Error; err != nil {
				return err
			}
		}

		var recipientIDs []string
		if err := tx.Model(&gormmodel.ConversationMember{}).Where("conversation_id = ?", message.ConversationID).
			Pluck("user_id", &recipientIDs).Error; err != nil {
			return err
		}
		storedText := ""
		if model.Text != nil {
			storedText = *model.Text
		}
		result.Message = entity.Message{ID: model.ID, ConversationID: model.ConversationID, SenderID: model.SenderID,
			Type: model.Type, Text: storedText, ClientMessageID: model.ClientMessageID, CreatedAt: model.CreatedAt,
			Attachment: message.Attachment}
		result.RecipientIDs = recipientIDs
		return nil
	})
	return result, err
}

func compareStoredAttachment(tx *gorm.DB, messageID string, expected *entity.MessageAttachment) error {
	var stored gormmodel.MessageAttachment
	err := tx.Where("message_id = ?", messageID).First(&stored).Error
	if expected == nil {
		if err == nil {
			return domainerrors.ErrConflict
		}
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if err != nil {
		return translate(err)
	}
	if stored.UploadID != expected.UploadID || stored.ObjectKey != expected.ObjectKey || stored.MIMEType != expected.MIMEType || stored.Size != expected.Size {
		return domainerrors.ErrConflict
	}
	expected.ID, expected.MessageID = stored.ID, stored.MessageID
	return nil
}
