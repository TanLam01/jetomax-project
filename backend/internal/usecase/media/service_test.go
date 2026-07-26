package media

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
)

type mediaRepositoryStub struct{ attachment *entity.MessageAttachment }

func (r *mediaRepositoryStub) Create(context.Context, *entity.MediaUpload) error { return nil }
func (r *mediaRepositoryStub) FindForUser(context.Context, string, string) (*entity.MediaUpload, error) {
	return nil, nil
}
func (r *mediaRepositoryStub) FindAttachmentForMember(context.Context, string, string) (*entity.MessageAttachment, error) {
	return r.attachment, nil
}

type storageStub struct{ downloadURL string }

func (s storageStub) PresignPut(context.Context, string, string, int64, time.Duration) (string, map[string]string, error) {
	return "", nil, nil
}
func (s storageStub) PresignGet(context.Context, string, time.Duration) (string, error) {
	return s.downloadURL, nil
}
func (s storageStub) VerifyObject(context.Context, string, string, int64) error { return nil }

func TestCreateDownloadForConversationMember(t *testing.T) {
	repository := &mediaRepositoryStub{attachment: &entity.MessageAttachment{ID: uuid.NewString(), ObjectKey: "users/u/images/a.jpg"}}
	service := NewService(repository, storageStub{downloadURL: "https://storage.example/signed"})
	download, err := service.CreateDownload(context.Background(), uuid.NewString(), repository.attachment.ID)
	if err != nil {
		t.Fatalf("create download: %v", err)
	}
	if download.URL != "https://storage.example/signed" || !download.ExpiresAt.After(time.Now()) {
		t.Fatalf("unexpected download: %+v", download)
	}
}
