package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/delivery/http/dto"
	mediausecase "github.com/jetomax/realtime-chat/backend/internal/usecase/media"
)

type MediaHandler struct{ service *mediausecase.Service }

func NewMediaHandler(service *mediausecase.Service) *MediaHandler {
	return &MediaHandler{service: service}
}

// CreateUpload godoc
// @Summary Create a pre-signed image upload
// @Description Accepts JPEG, PNG or WebP metadata up to 10 MB and returns a pre-signed S3/MinIO PUT URL valid for 15 minutes.
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateMediaUploadRequest true "Image metadata"
// @Success 201 {object} dto.CreateMediaUploadResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /media/uploads [post]
func (h *MediaHandler) CreateUpload(c *gin.Context) {
	var request dto.CreateMediaUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	upload, err := h.service.CreateUpload(c.Request.Context(), authenticatedUserID(c), request.FileName, request.MIMEType, request.Size)
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewCreateMediaUploadResponse(upload))
}

// CreateDownload godoc
// @Summary Create a signed image download URL
// @Description Returns a short-lived URL only when the authenticated user is a member of the attachment conversation.
// @Tags media
// @Produce json
// @Security BearerAuth
// @Param id path string true "Attachment UUID"
// @Success 200 {object} dto.MediaDownloadResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /media/attachments/{id}/download [get]
func (h *MediaHandler) CreateDownload(c *gin.Context) {
	download, err := h.service.CreateDownload(c.Request.Context(), authenticatedUserID(c), c.Param("id"))
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewMediaDownloadResponse(download))
}
