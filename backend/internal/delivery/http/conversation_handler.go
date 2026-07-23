package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/delivery/http/dto"
	conversationusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/conversation"
)

type ConversationHandler struct{ service *conversationusecase.Service }

func NewConversationHandler(service *conversationusecase.Service) *ConversationHandler {
	return &ConversationHandler{service: service}
}

// List godoc
// @Summary List conversations for the authenticated member
// @Description Returns up to 50 conversations ordered by latest activity with last message and unread count.
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ConversationListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /conversations [get]
func (h *ConversationHandler) List(c *gin.Context) {
	conversations, err := h.service.List(c.Request.Context(), authenticatedUserID(c))
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewConversationListResponse(conversations))
}

// CreateDirect godoc
// @Summary Create or get a direct conversation
// @Description Returns 201 when created and 200 when the same user pair already has a direct conversation.
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateDirectRequest true "Target user"
// @Success 200 {object} dto.ConversationResponse
// @Success 201 {object} dto.ConversationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /conversations/direct [post]
func (h *ConversationHandler) CreateDirect(c *gin.Context) {
	var request dto.CreateDirectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	conversation, created, err := h.service.CreateDirect(c.Request.Context(), authenticatedUserID(c), request.UserID)
	if err != nil {
		resourceError(c, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, dto.NewConversationResponse(*conversation))
}

// CreateGroup godoc
// @Summary Create a group conversation
// @Description The authenticated creator becomes owner; invited users become members.
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGroupRequest true "Group details"
// @Success 201 {object} dto.ConversationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /conversations/groups [post]
func (h *ConversationHandler) CreateGroup(c *gin.Context) {
	var request dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	conversation, err := h.service.CreateGroup(c.Request.Context(), authenticatedUserID(c), request.Name, request.AvatarKey, request.MemberIDs)
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewConversationResponse(*conversation))
}
