package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jetomax/realtime-chat/backend/internal/delivery/http/dto"
	domainerrors "github.com/jetomax/realtime-chat/backend/internal/domain/errors"
	authusecase "github.com/jetomax/realtime-chat/backend/internal/usecase/auth"
)

type AuthHandler struct{ service *authusecase.Service }

func NewAuthHandler(service *authusecase.Service) *AuthHandler { return &AuthHandler{service: service} }

// Register godoc
// @Summary Register a user
// @Description Creates an account. The password is stored only as a one-way bcrypt hash.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration payload"
// @Success 201 {object} dto.SessionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var request dto.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	session, err := h.service.Register(c.Request.Context(), request.Email, request.DisplayName, request.Password)
	if err != nil {
		authError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewSessionResponse(session))
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.SessionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	session, err := h.service.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		authError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewSessionResponse(session))
}

// Refresh godoc
// @Summary Rotate a refresh token
// @Description Revokes the supplied refresh token and issues a new session.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.SessionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var request dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	session, err := h.service.Refresh(c.Request.Context(), request.RefreshToken)
	if err != nil {
		authError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewSessionResponse(session))
}

// Logout godoc
// @Summary Logout
// @Description Revokes the supplied refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var request dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		validationError(c, "invalid request body")
		return
	}
	if err := h.service.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		authError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func authError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainerrors.ErrValidation):
		respondError(c, http.StatusBadRequest, "validation_error", err.Error())
	case errors.Is(err, domainerrors.ErrConflict):
		respondError(c, http.StatusConflict, "conflict", "email is already registered")
	case errors.Is(err, domainerrors.ErrUnauthorized), errors.Is(err, domainerrors.ErrNotFound):
		respondError(c, http.StatusUnauthorized, "unauthorized", "invalid credentials or token")
	default:
		respondError(c, http.StatusInternalServerError, "internal_error", "an internal error occurred")
	}
}

func validationError(c *gin.Context, message string) {
	respondError(c, http.StatusBadRequest, "validation_error", message)
}

func respondError(c *gin.Context, status int, code, message string) {
	setSafeError(c, code, message)
	c.JSON(status, dto.NewPublicErrorResponse(status))
}
