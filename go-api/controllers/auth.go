package controllers

import (
	"errors"
	"net/http"
	"poc-gin/pkg/logger"
	"poc-gin/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	user, err := h.authService.Register(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyUsed) {
			RespondError(c, http.StatusConflict, "EMAIL_ALREADY_USED", "Email already registered", nil)
			return
		}
		logger.Error("Failed to register user: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user", nil)
		return
	}

	RespondSuccess(c, http.StatusCreated, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	token, refreshToken, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
			return
		}
		logger.Error("Login error: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Login failed", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, gin.H{"token": token, "refresh_token": refreshToken})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	token, refreshToken, err := h.authService.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) ||
			errors.Is(err, services.ErrRefreshTokenExpired) ||
			errors.Is(err, services.ErrRefreshTokenReused) {
			RespondError(c, http.StatusUnauthorized, "REFRESH_TOKEN_INVALID", "Invalid or expired refresh token", nil)
			return
		}
		logger.Error("Refresh error: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Refresh failed", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, gin.H{"token": token, "refresh_token": refreshToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	if err := h.authService.Logout(ctx, req.RefreshToken); err != nil {
		logger.Error("Logout error: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Logout failed", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, gin.H{"success": true})
}
