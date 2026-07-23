package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/delivery/http/dto"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	userusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/user"
)

type UserHandler struct{ service *userusecase.Service }

func NewUserHandler(service *userusecase.Service) *UserHandler { return &UserHandler{service: service} }

// Me godoc
// @Summary Get the authenticated user profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	user, err := h.service.Me(c.Request.Context(), authenticatedUserID(c))
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewUserResponse(*user))
}

// Search godoc
// @Summary Search users by display name or email
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param query query string true "Search query (2-100 characters)"
// @Success 200 {object} dto.UserListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h *UserHandler) Search(c *gin.Context) {
	users, err := h.service.Search(c.Request.Context(), authenticatedUserID(c), c.Query("query"))
	if err != nil {
		resourceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewUserListResponse(users))
}

func resourceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrValidation):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, domainerrors.ErrNotFound):
		respondError(c, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domainerrors.ErrForbidden):
		respondError(c, http.StatusForbidden, "conversation_access_denied", "user is not a conversation member")
	default:
		respondError(c, http.StatusInternalServerError, "internal_error", "an internal error occurred")
	}
}
