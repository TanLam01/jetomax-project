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
