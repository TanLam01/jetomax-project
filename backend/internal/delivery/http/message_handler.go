package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/delivery/http/dto"
	"github.com/jetomax/realtime-chat/backend/internal/domain/entity"
	messageusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/message"
)

type MessageHandler struct{ service *messageusecase.Service }

func NewMessageHandler(service *messageusecase.Service) *MessageHandler {
	return &MessageHandler{service: service}
}

type messageCursor struct {
	CreatedAt string `json:"t"`
	MessageID string `json:"i"`
}

// List godoc
// @Summary Get conversation message history
// @Description Returns messages newest-first using cursor pagination. Only conversation members may access this endpoint.
// @Tags messages
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param cursor query string false "Opaque cursor returned by the previous response"
// @Param limit query int false "Page size (default 20, maximum 100)"
// @Success 200 {object} dto.MessagePageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /conversations/{id}/messages [get]
func (h *MessageHandler) List(c *gin.Context) {
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			validationError(c, "limit must be an integer")
			return
		}
		if parsed < 1 || parsed > 100 {
			validationError(c, "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}
	cursor, err := decodeMessageCursor(c.Query("cursor"))
	if err != nil {
		validationError(c, err.Error())
		return
	}
	page, err := h.service.List(c.Request.Context(), authenticatedUserID(c), c.Param("id"), cursor, limit)
	if err != nil {
		resourceError(c, err)
		return
	}
	nextCursor := ""
	if page.HasMore && len(page.Messages) > 0 {
		nextCursor, err = encodeMessageCursor(page.Messages[len(page.Messages)-1])
		if err != nil {
			resourceError(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, dto.NewMessagePageResponse(page.Messages, page.HasMore, nextCursor))
}

func decodeMessageCursor(raw string) (*messageusecase.Cursor, error) {
	if raw == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor")
	}
	var value messageCursor
	if err := json.Unmarshal(decoded, &value); err != nil {
		return nil, fmt.Errorf("invalid cursor")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, value.CreatedAt)
	if err != nil || value.MessageID == "" {
		return nil, fmt.Errorf("invalid cursor")
	}
	return &messageusecase.Cursor{CreatedAt: createdAt, MessageID: value.MessageID}, nil
}

func encodeMessageCursor(message entity.Message) (string, error) {
	encoded, err := json.Marshal(messageCursor{CreatedAt: message.CreatedAt.Format(time.RFC3339Nano), MessageID: message.ID})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}
