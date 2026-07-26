package message

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type messageRepositoryStub struct {
	received   entity.Message
	recipients []string
	markedRead string
}

func (r *messageRepositoryStub) ListForMember(context.Context, string, string, *time.Time, string, int) ([]entity.Message, error) {
	return nil, nil
}

func (r *messageRepositoryStub) ListAfterForMember(context.Context, string, string, time.Time, string, int) ([]entity.Message, error) {
	return nil, nil
}

func (r *messageRepositoryStub) RecipientIDsForMember(context.Context, string, string) ([]string, error) {
	return r.recipients, nil
}

func (r *messageRepositoryStub) PeerIDsForUser(context.Context, string) ([]string, error) {
	return nil, nil
}

func (r *messageRepositoryStub) MarkRead(_ context.Context, _, _, messageID string) error {
	r.markedRead = messageID
	return nil
}

func (r *messageRepositoryStub) Send(_ context.Context, message entity.Message) (*entity.SendMessageResult, error) {
	r.received = message
	message.ID = uuid.NewString()
	message.CreatedAt = time.Now().UTC()
	return &entity.SendMessageResult{Message: message, RecipientIDs: []string{uuid.NewString()}, Created: true}, nil
}

type publisherStub struct {
	event          entity.MessageEvent
	realtimeEvents []entity.RealtimeEvent
}

func (p *publisherStub) PublishMessage(_ context.Context, event entity.MessageEvent) error {
	p.event = event
	return nil
}

func (p *publisherStub) PublishRealtime(_ context.Context, event entity.RealtimeEvent) error {
	p.realtimeEvents = append(p.realtimeEvents, event)
	return nil
}

type imageResolverStub struct{ attachment *entity.MessageAttachment }

func (r imageResolverStub) ResolveImage(context.Context, string, string) (*entity.MessageAttachment, error) {
	return r.attachment, nil
}

func TestSendImageIncludesVerifiedAttachment(t *testing.T) {
	repository := &messageRepositoryStub{}
	publisher := &publisherStub{}
	attachment := &entity.MessageAttachment{UploadID: uuid.NewString(), ObjectKey: "users/user/images/image.jpg", MIMEType: "image/jpeg", Size: 128}
	service := NewService(repository, publisher, imageResolverStub{attachment: attachment})

	message, err := service.Send(context.Background(), uuid.NewString(), SendInput{
		ConversationID: uuid.NewString(), ClientMessageID: uuid.NewString(), Type: "image",
		Text: "caption", UploadID: attachment.UploadID,
	})
	if err != nil {
		t.Fatalf("send image: %v", err)
	}
	if repository.received.Attachment != attachment {
		t.Fatal("verified attachment was not passed to repository")
	}
	if message.Attachment != attachment || publisher.event.Message.Attachment != attachment {
		t.Fatal("attachment was not included in stored/published message")
	}
}

func TestSendImageRequiresUploadID(t *testing.T) {
	service := NewService(&messageRepositoryStub{}, &publisherStub{}, imageResolverStub{})
	_, err := service.Send(context.Background(), uuid.NewString(), SendInput{
		ConversationID: uuid.NewString(), ClientMessageID: uuid.NewString(), Type: "image",
	})
	if err == nil {
		t.Fatal("expected image without upload_id to fail")
	}
}

func TestPublishTypingExcludesSender(t *testing.T) {
	senderID, recipientID := uuid.NewString(), uuid.NewString()
	repository := &messageRepositoryStub{recipients: []string{senderID, recipientID}}
	publisher := &publisherStub{}
	service := NewService(repository, publisher, imageResolverStub{})
	if err := service.PublishTyping(context.Background(), senderID, uuid.NewString(), "request-1", true); err != nil {
		t.Fatalf("publish typing: %v", err)
	}
	if len(publisher.realtimeEvents) != 1 || publisher.realtimeEvents[0].Type != "typing.changed" {
		t.Fatalf("unexpected realtime events: %+v", publisher.realtimeEvents)
	}
	if len(publisher.realtimeEvents[0].RecipientIDs) != 1 || publisher.realtimeEvents[0].RecipientIDs[0] != recipientID {
		t.Fatalf("unexpected typing recipients: %v", publisher.realtimeEvents[0].RecipientIDs)
	}
}

func TestMarkReadPublishesConversationUpdate(t *testing.T) {
	userID, messageID := uuid.NewString(), uuid.NewString()
	repository := &messageRepositoryStub{recipients: []string{userID}}
	publisher := &publisherStub{}
	service := NewService(repository, publisher, imageResolverStub{})
	if err := service.MarkRead(context.Background(), userID, uuid.NewString(), messageID, "request-2"); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if repository.markedRead != messageID || len(publisher.realtimeEvents) != 1 || publisher.realtimeEvents[0].Type != "conversation.updated" {
		t.Fatalf("read update was not stored/published: marked=%s events=%+v", repository.markedRead, publisher.realtimeEvents)
	}
}
