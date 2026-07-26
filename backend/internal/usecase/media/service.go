package media

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	"github.com/jetomax/realtime-chat/backend/internal/domain/repository"
)

const maxImageSize int64 = 10 * 1024 * 1024

var imageExtensions = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

type Service struct {
	repository  repository.MediaUploadRepository
	attachments repository.MediaAttachmentRepository
	signer      repository.UploadSigner
	now         func() time.Time
}

type Upload struct {
	entity.MediaUpload
	UploadURL string
	Headers   map[string]string
}

func NewService(uploadRepository repository.MediaUploadRepository, signer repository.UploadSigner) *Service {
	attachments, _ := uploadRepository.(repository.MediaAttachmentRepository)
	return &Service{repository: uploadRepository, attachments: attachments, signer: signer, now: time.Now}
}

type Download struct {
	URL       string
	ExpiresAt time.Time
}

func (s *Service) CreateDownload(ctx context.Context, userID, attachmentID string) (*Download, error) {
	if _, err := uuid.Parse(attachmentID); err != nil {
		return nil, fmt.Errorf("%w: invalid attachment id", domainerrors.ErrValidation)
	}
	if s.attachments == nil {
		return nil, fmt.Errorf("media attachment repository is not configured")
	}
	attachment, err := s.attachments.FindAttachmentForMember(ctx, attachmentID, userID)
	if err != nil {
		return nil, err
	}
	signer, ok := s.signer.(repository.DownloadSigner)
	if !ok {
		return nil, fmt.Errorf("storage does not support signed downloads")
	}
	ttl := 15 * time.Minute
	url, err := signer.PresignGet(ctx, attachment.ObjectKey, ttl)
	if err != nil {
		return nil, err
	}
	return &Download{URL: url, ExpiresAt: s.now().UTC().Add(ttl)}, nil
}

func (s *Service) CreateUpload(ctx context.Context, userID, fileName, mimeType string, size int64) (*Upload, error) {
	fileName, mimeType = filepath.Base(strings.TrimSpace(fileName)), strings.ToLower(strings.TrimSpace(mimeType))
	extension, allowed := imageExtensions[mimeType]
	if !allowed {
		return nil, fmt.Errorf("%w: unsupported image MIME type", domainerrors.ErrValidation)
	}
	if fileName == "." || fileName == "" || len(fileName) > 255 {
		return nil, fmt.Errorf("%w: invalid file name", domainerrors.ErrValidation)
	}
	if size <= 0 || size > maxImageSize {
		return nil, fmt.Errorf("%w: image size must be between 1 byte and 10 MB", domainerrors.ErrValidation)
	}
	now, ttl := s.now().UTC(), 15*time.Minute
	id := uuid.NewString()
	objectKey := fmt.Sprintf("users/%s/images/%s%s", userID, id, extension)
	url, headers, err := s.signer.PresignPut(ctx, objectKey, mimeType, size, ttl)
	if err != nil {
		return nil, err
	}
	upload := entity.MediaUpload{ID: id, UserID: userID, ObjectKey: objectKey, OriginalName: fileName,
		MIMEType: mimeType, Size: size, Status: "pending", ExpiresAt: now.Add(ttl)}
	if err := s.repository.Create(ctx, &upload); err != nil {
		return nil, err
	}
	return &Upload{MediaUpload: upload, UploadURL: url, Headers: headers}, nil
}

func (s *Service) ResolveImage(ctx context.Context, userID, uploadID string) (*entity.MessageAttachment, error) {
	if _, err := uuid.Parse(uploadID); err != nil {
		return nil, fmt.Errorf("%w: invalid upload_id", domainerrors.ErrValidation)
	}
	upload, err := s.repository.FindForUser(ctx, uploadID, userID)
	if err != nil {
		return nil, err
	}
	if upload.Status != "pending" && upload.Status != "uploaded" {
		return nil, fmt.Errorf("%w: media upload is not available", domainerrors.ErrConflict)
	}
	if upload.Status == "pending" && !s.now().UTC().Before(upload.ExpiresAt) {
		return nil, fmt.Errorf("%w: media upload has expired", domainerrors.ErrConflict)
	}
	inspector, ok := s.signer.(repository.UploadInspector)
	if !ok {
		return nil, fmt.Errorf("upload storage does not support object verification")
	}
	if err := inspector.VerifyObject(ctx, upload.ObjectKey, upload.MIMEType, upload.Size); err != nil {
		return nil, fmt.Errorf("%w: uploaded image is missing or does not match metadata: %v", domainerrors.ErrConflict, err)
	}
	return &entity.MessageAttachment{UploadID: upload.ID, ObjectKey: upload.ObjectKey,
		MIMEType: upload.MIMEType, Size: upload.Size}, nil
}
