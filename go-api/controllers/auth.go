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
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", err.Error())
		return
	}

	user, err := h.authService.Register(req.Email, req.Password)
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
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", err.Error())
		return
	}

	token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
			return
		}
		logger.Error("Login error: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Login failed", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, gin.H{"token": token})
}
