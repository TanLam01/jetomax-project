package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/infrastructure/persistence/gormmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Conversation struct{ db *gorm.DB }

func NewConversation(db *gorm.DB) *Conversation { return &Conversation{db: db} }

type conversationRow struct {
	ID                   string
	Type                 string
	Name                 sql.NullString
	AvatarKey            sql.NullString
	CreatedBy            string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Role                 string
	UnreadCount          int64
	LastMessageID        sql.NullString
	LastSenderID         sql.NullString
	LastMessageType      sql.NullString
	LastMessageText      sql.NullString
	LastClientMessageID  sql.NullString
	LastMessageCreatedAt sql.NullTime
}

func (r *Conversation) CreateDirect(ctx context.Context, creatorID, targetID, directKey string) (*entity.ConversationSummary, bool, error) {
	var result entity.ConversationSummary
	created := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCount int64
		if err := tx.Model(&gormmodel.User{}).Where("id IN ?", []string{creatorID, targetID}).Count(&userCount).Error; err != nil {
			return err
		}
		if userCount != 2 {
			return domainerrors.ErrNotFound
		}
		model := gormmodel.Conversation{Type: "direct", DirectKey: &directKey, CreatedBy: creatorID}
		createResult := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "direct_key"}}, DoNothing: true}).Create(&model)
		if createResult.Error != nil {
			return translate(createResult.Error)
		}
		created = createResult.RowsAffected == 1
		if !created {
			if err := tx.Where("direct_key = ?", directKey).First(&model).Error; err != nil {
				return translate(err)
			}
		} else {
			members := []gormmodel.ConversationMember{
				{ConversationID: model.ID, UserID: creatorID, Role: "member"},
				{ConversationID: model.ID, UserID: targetID, Role: "member"},
			}
			if err := tx.Create(&members).Error; err != nil {
				return translate(err)
			}
		}
		result = entity.ConversationSummary{Conversation: conversationToEntity(model), Role: "member"}
		return nil
	})
	return &result, created, err
}

func (r *Conversation) CreateGroup(ctx context.Context, creatorID string, input entity.CreateGroupInput) (*entity.ConversationSummary, error) {
	var result entity.ConversationSummary
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		memberIDs := append([]string{creatorID}, input.MemberIDs...)
		var count int64
		if err := tx.Model(&gormmodel.User{}).Where("id IN ?", memberIDs).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(memberIDs)) {
			return domainerrors.ErrNotFound
		}
		name := input.Name
		var avatar *string
		if input.AvatarKey != "" {
			avatar = &input.AvatarKey
		}
		model := gormmodel.Conversation{Type: "group", Name: &name, AvatarKey: avatar, CreatedBy: creatorID}
		if err := tx.Create(&model).Error; err != nil {
			return translate(err)
		}
		members := make([]gormmodel.ConversationMember, 0, len(memberIDs))
		members = append(members, gormmodel.ConversationMember{ConversationID: model.ID, UserID: creatorID, Role: "owner"})
		for _, memberID := range input.MemberIDs {
			members = append(members, gormmodel.ConversationMember{ConversationID: model.ID, UserID: memberID, Role: "member"})
		}
		if err := tx.Create(&members).Error; err != nil {
			return translate(err)
		}
		result = entity.ConversationSummary{Conversation: conversationToEntity(model), Role: "owner"}
		return nil
	})
	return &result, err
}

func (r *Conversation) AddMembers(ctx context.Context, conversationID, actorID string, userIDs []string) ([]string, error) {
	added := make([]string, 0, len(userIDs))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var conversation gormmodel.Conversation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&conversation, "id = ?", conversationID).Error; err != nil {
			return translate(err)
		}
		if conversation.Type != "group" {
			return fmt.Errorf("%w: members can only be added to group conversations", domainerrors.ErrValidation)
		}
		var actor gormmodel.ConversationMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&actor, "conversation_id = ? AND user_id = ?", conversationID, actorID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerrors.ErrForbidden
			}
			return err
		}
		if actor.Role != "owner" && actor.Role != "admin" {
			return domainerrors.ErrForbidden
		}

		var userCount int64
		if err := tx.Model(&gormmodel.User{}).Where("id IN ?", userIDs).Count(&userCount).Error; err != nil {
			return err
		}
		if userCount != int64(len(userIDs)) {
			return domainerrors.ErrNotFound
		}
		for _, userID := range userIDs {
			member := gormmodel.ConversationMember{ConversationID: conversationID, UserID: userID, Role: "member"}
			result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&member)
			if result.Error != nil {
				return translate(result.Error)
			}
			if result.RowsAffected == 1 {
				added = append(added, userID)
			}
		}
		return nil
	})
	return added, err
}

func conversationToEntity(model gormmodel.Conversation) entity.Conversation {
	name, avatar := "", ""
	if model.Name != nil {
		name = *model.Name
	}
	if model.AvatarKey != nil {
		avatar = *model.AvatarKey
	}
	return entity.Conversation{ID: model.ID, Type: model.Type, Name: name, AvatarKey: avatar,
		CreatedBy: model.CreatedBy, CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt}
}

func (r *Conversation) ListForUser(ctx context.Context, userID string, limit int) ([]entity.ConversationSummary, error) {
	var rows []conversationRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.id, c.type::text AS type, c.name, c.avatar_key, c.created_by,
		       c.created_at, c.updated_at, cm.role::text AS role,
		       COALESCE((
		           SELECT COUNT(*) FROM messages unread
		           WHERE unread.conversation_id = c.id
		             AND unread.sender_id <> ?
		             AND (cm.last_read_message_id IS NULL OR unread.created_at > COALESCE(
		                 (SELECT read_message.created_at FROM messages read_message WHERE read_message.id = cm.last_read_message_id),
		                 '-infinity'::timestamptz
		             ))
		       ), 0) AS unread_count,
		       last_message.id AS last_message_id, last_message.sender_id AS last_sender_id,
		       last_message.type::text AS last_message_type, last_message.text AS last_message_text,
		       last_message.client_message_id AS last_client_message_id,
		       last_message.created_at AS last_message_created_at
		FROM conversation_members cm
		JOIN conversations c ON c.id = cm.conversation_id
		LEFT JOIN LATERAL (
		    SELECT m.* FROM messages m WHERE m.conversation_id = c.id
		    ORDER BY m.created_at DESC, m.id DESC LIMIT 1
		) last_message ON true
		WHERE cm.user_id = ?
		ORDER BY COALESCE(last_message.created_at, c.updated_at) DESC, c.id DESC
		LIMIT ?`, userID, userID, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]entity.ConversationSummary, 0, len(rows))
	for _, row := range rows {
		summary := entity.ConversationSummary{Conversation: entity.Conversation{
			ID: row.ID, Type: row.Type, Name: row.Name.String, AvatarKey: row.AvatarKey.String,
			CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}, Role: row.Role, UnreadCount: row.UnreadCount}
		if row.LastMessageID.Valid {
			summary.LastMessage = &entity.Message{ID: row.LastMessageID.String, ConversationID: row.ID,
				SenderID: row.LastSenderID.String, Type: row.LastMessageType.String, Text: row.LastMessageText.String,
				ClientMessageID: row.LastClientMessageID.String, CreatedAt: row.LastMessageCreatedAt.Time}
		}
		result = append(result, summary)
	}
	return result, nil
}
