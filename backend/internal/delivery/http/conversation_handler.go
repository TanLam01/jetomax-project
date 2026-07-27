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

// AddMembers godoc
// @Summary Add members to a group conversation
// @Description Only the group owner or an admin can add users. Existing members are ignored.
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param request body dto.AddMembersRequest true "Users to add"
// @Success 200 {object} dto.AddMembersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /conversations/{id}/members [post]
func (h *ConversationHandler) AddMembers(c *gin.Context) {
	var request dto.AddMembersRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	added, err := h.service.AddMembers(c.Request.Context(), authenticatedUserID(c), c.Param("id"), request.UserIDs)
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewAddMembersResponse(added))
}

// RemoveMember godoc
// @Summary Remove a member from a group or leave a group
// @Description Owners can remove admins/members, admins can remove members, and non-owner users can remove themselves to leave.
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param userId path string true "Target user UUID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /conversations/{id}/members/{userId} [delete]
func (h *ConversationHandler) RemoveMember(c *gin.Context) {
	err := h.service.RemoveMember(c.Request.Context(), authenticatedUserID(c), c.Param("id"), c.Param("userId"))
	if err != nil {
		resourceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListMembers godoc
// @Summary List group members
// @Description Any group member may list member profiles and roles.
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Success 200 {object} dto.ConversationMemberListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /conversations/{id}/members [get]
func (h *ConversationHandler) ListMembers(c *gin.Context) {
	members, err := h.service.ListMembers(c.Request.Context(), authenticatedUserID(c), c.Param("id"))
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewConversationMemberListResponse(members))
}

// UpdateMemberRole godoc
// @Summary Promote or demote a group member
// @Description Only the owner may change another member between admin and member.
// @Tags conversations
// @Accept json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param userId path string true "Target user UUID"
// @Param request body dto.UpdateMemberRoleRequest true "New role"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /conversations/{id}/members/{userId}/role [patch]
func (h *ConversationHandler) UpdateMemberRole(c *gin.Context) {
	var request dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	if err := h.service.UpdateMemberRole(c.Request.Context(), authenticatedUserID(c), c.Param("id"), c.Param("userId"), request.Role); err != nil {
		resourceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// TransferOwnership godoc
// @Summary Transfer group ownership
// @Description Only the current owner may transfer ownership to another member; the previous owner becomes admin.
// @Tags conversations
// @Accept json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param request body dto.TransferOwnershipRequest true "New owner"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /conversations/{id}/ownership [post]
func (h *ConversationHandler) TransferOwnership(c *gin.Context) {
	var request dto.TransferOwnershipRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	if err := h.service.TransferOwnership(c.Request.Context(), authenticatedUserID(c), c.Param("id"), request.UserID); err != nil {
		resourceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
