package message

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
)

type Cursor struct {
	CreatedAt time.Time
	MessageID string
}

type Page struct {
	Messages []entity.Message
	HasMore  bool
}

type Service struct {
	repository repository.MessageRepository
	publisher  repository.MessagePublisher
	images     ImageResolver
	realtime   repository.RealtimePublisher
	presence   repository.PresenceTracker
}

type ImageResolver interface {
	ResolveImage(context.Context, string, string) (*entity.MessageAttachment, error)
}

func NewService(messageRepository repository.MessageRepository, publisher repository.MessagePublisher, images ImageResolver) *Service {
	service := &Service{repository: messageRepository, publisher: publisher, images: images}
	service.realtime, _ = publisher.(repository.RealtimePublisher)
	service.presence, _ = publisher.(repository.PresenceTracker)
	return service
}

type SendInput struct{ ConversationID, ClientMessageID, Type, Text, UploadID string }

func (s *Service) Send(ctx context.Context, senderID string, input SendInput) (*entity.Message, error) {
	if _, err := uuid.Parse(input.ConversationID); err != nil {
		return nil, fmt.Errorf("%w: invalid conversation id", domainerrors.ErrValidation)
	}
	if _, err := uuid.Parse(input.ClientMessageID); err != nil {
		return nil, fmt.Errorf("%w: invalid client_message_id", domainerrors.ErrValidation)
	}
	input.Text = strings.TrimSpace(input.Text)
	var attachment *entity.MessageAttachment
	switch input.Type {
	case "text":
		if input.Text == "" || len([]rune(input.Text)) > 4000 {
			return nil, fmt.Errorf("%w: text must contain 1-4000 characters", domainerrors.ErrValidation)
		}
		if input.UploadID != "" {
			return nil, fmt.Errorf("%w: text message must not contain upload_id", domainerrors.ErrValidation)
		}
	case "image":
		if input.UploadID == "" {
			return nil, fmt.Errorf("%w: image message requires upload_id", domainerrors.ErrValidation)
		}
		if len([]rune(input.Text)) > 4000 {
			return nil, fmt.Errorf("%w: image caption must not exceed 4000 characters", domainerrors.ErrValidation)
		}
		if s.images == nil {
			return nil, fmt.Errorf("image resolver is not configured")
		}
		var err error
		attachment, err = s.images.ResolveImage(ctx, senderID, input.UploadID)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: unsupported message type", domainerrors.ErrValidation)
	}
	result, err := s.repository.Send(ctx, entity.Message{ConversationID: input.ConversationID, SenderID: senderID,
		Type: input.Type, Text: input.Text, ClientMessageID: input.ClientMessageID, Attachment: attachment})
	if err != nil {
		return nil, err
	}
	// Publish retries too. If PostgreSQL committed but Redis was temporarily unavailable,
	// retrying the same client_message_id must be able to deliver the stored message.
	if s.publisher != nil {
		event := entity.MessageEvent{EventID: uuid.NewString(), RecipientIDs: result.RecipientIDs, Message: result.Message, Timestamp: time.Now().UTC()}
		if err := s.publisher.PublishMessage(ctx, event); err != nil {
			return nil, fmt.Errorf("publish message: %w", err)
		}
		if s.realtime != nil {
			updated := entity.RealtimeEvent{Type: "conversation.updated", EventID: uuid.NewString(), RecipientIDs: result.RecipientIDs,
				Timestamp: time.Now().UTC(), Payload: map[string]any{"conversation_id": result.Message.ConversationID,
					"last_message_id": result.Message.ID, "updated_at": result.Message.CreatedAt.UTC().Format(time.RFC3339Nano)}}
			if err := s.realtime.PublishRealtime(ctx, updated); err != nil {
				slog.Error("publish conversation updated event", "conversation_id", result.Message.ConversationID, "error", err)
			}
		}
	}
	return &result.Message, nil
}

func (s *Service) PublishTyping(ctx context.Context, userID, conversationID, requestID string, isTyping bool) error {
	if _, err := uuid.Parse(conversationID); err != nil {
		return fmt.Errorf("%w: invalid conversation id", domainerrors.ErrValidation)
	}
	recipients, err := s.repository.RecipientIDsForMember(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	recipients = excluding(recipients, userID)
	if s.realtime == nil || len(recipients) == 0 {
		return nil
	}
	return s.realtime.PublishRealtime(ctx, entity.RealtimeEvent{Type: "typing.changed", EventID: uuid.NewString(),
		RequestID: requestID, RecipientIDs: recipients, Timestamp: time.Now().UTC(),
		Payload: map[string]any{"conversation_id": conversationID, "user_id": userID, "is_typing": isTyping}})
}

func (s *Service) MarkRead(ctx context.Context, userID, conversationID, messageID, requestID string) error {
	if _, err := uuid.Parse(conversationID); err != nil {
		return fmt.Errorf("%w: invalid conversation id", domainerrors.ErrValidation)
	}
	if _, err := uuid.Parse(messageID); err != nil {
		return fmt.Errorf("%w: invalid message id", domainerrors.ErrValidation)
	}
	if err := s.repository.MarkRead(ctx, conversationID, userID, messageID); err != nil {
		return err
	}
	recipients, err := s.repository.RecipientIDsForMember(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if s.realtime == nil {
		return nil
	}
	return s.realtime.PublishRealtime(ctx, entity.RealtimeEvent{Type: "conversation.updated", EventID: uuid.NewString(),
		RequestID: requestID, RecipientIDs: recipients, Timestamp: time.Now().UTC(),
		Payload: map[string]any{"conversation_id": conversationID, "read_by_user_id": userID, "last_read_message_id": messageID}})
}

func (s *Service) PresenceConnected(ctx context.Context, userID, connectionID string) error {
	if s.presence == nil || s.realtime == nil {
		return nil
	}
	becameOnline, err := s.presence.ConnectPresence(ctx, userID, connectionID)
	if err != nil || !becameOnline {
		return err
	}
	return s.publishPresence(ctx, userID, true)
}

func (s *Service) TouchPresence(ctx context.Context, userID, connectionID string) error {
	if s.presence == nil {
		return nil
	}
	return s.presence.TouchPresence(ctx, userID, connectionID)
}

func (s *Service) PresenceDisconnected(ctx context.Context, userID, connectionID string) error {
	if s.presence == nil || s.realtime == nil {
		return nil
	}
	becameOffline, err := s.presence.DisconnectPresence(ctx, userID, connectionID)
	if err != nil || !becameOffline {
		return err
	}
	return s.publishPresence(ctx, userID, false)
}

func (s *Service) OnlinePeers(ctx context.Context, userID string) ([]string, error) {
	if s.presence == nil {
		return nil, nil
	}
	peers, err := s.repository.PeerIDsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	states, err := s.presence.OnlinePresence(ctx, peers)
	if err != nil {
		return nil, err
	}
	online := make([]string, 0, len(peers))
	for _, peerID := range peers {
		if states[peerID] {
			online = append(online, peerID)
		}
	}
	return online, nil
}

func (s *Service) publishPresence(ctx context.Context, userID string, online bool) error {
	recipients, err := s.repository.PeerIDsForUser(ctx, userID)
	if err != nil || len(recipients) == 0 {
		return err
	}
	return s.realtime.PublishRealtime(ctx, entity.RealtimeEvent{Type: "presence.changed", EventID: uuid.NewString(),
		RecipientIDs: recipients, Timestamp: time.Now().UTC(), Payload: map[string]any{"user_id": userID, "online": online}})
}

func excluding(ids []string, excluded string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != excluded {
			result = append(result, id)
		}
	}
	return result
}

func (s *Service) List(ctx context.Context, userID, conversationID string, cursor *Cursor, limit int) (*Page, error) {
	if _, err := uuid.Parse(conversationID); err != nil {
		return nil, fmt.Errorf("%w: invalid conversation id", domainerrors.ErrValidation)
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		return nil, fmt.Errorf("%w: limit must not exceed 100", domainerrors.ErrValidation)
	}
	var before *time.Time
	beforeID := ""
	if cursor != nil {
		if cursor.CreatedAt.IsZero() {
			return nil, fmt.Errorf("%w: invalid cursor timestamp", domainerrors.ErrValidation)
		}
		if _, err := uuid.Parse(cursor.MessageID); err != nil {
			return nil, fmt.Errorf("%w: invalid cursor message id", domainerrors.ErrValidation)
		}
		before, beforeID = &cursor.CreatedAt, cursor.MessageID
	}
	messages, err := s.repository.ListForMember(ctx, conversationID, userID, before, beforeID, limit+1)
	if err != nil {
		return nil, err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	return &Page{Messages: messages, HasMore: hasMore}, nil
}

func (s *Service) SyncAfter(ctx context.Context, userID, conversationID string, cursor Cursor, limit int) (*Page, error) {
	if _, err := uuid.Parse(conversationID); err != nil {
		return nil, fmt.Errorf("%w: invalid conversation id", domainerrors.ErrValidation)
	}
	if cursor.CreatedAt.IsZero() {
		return nil, fmt.Errorf("%w: invalid cursor timestamp", domainerrors.ErrValidation)
	}
	if _, err := uuid.Parse(cursor.MessageID); err != nil {
		return nil, fmt.Errorf("%w: invalid cursor message id", domainerrors.ErrValidation)
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		return nil, fmt.Errorf("%w: limit must not exceed 100", domainerrors.ErrValidation)
	}
	messages, err := s.repository.ListAfterForMember(ctx, conversationID, userID, cursor.CreatedAt, cursor.MessageID, limit+1)
	if err != nil {
		return nil, err
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	return &Page{Messages: messages, HasMore: hasMore}, nil
}
