package controllers

import (
	"errors"
	"net/http"
	"poc-gin/pkg/logger"
	"poc-gin/services"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService services.ProfileServiceInterface
}

func NewProfileHandler(profileService services.ProfileServiceInterface) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", errorInvalidAuthenticationContext, nil)
		return
	}

	stats, err := h.profileService.GetProfileStats(ctx, userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			RespondError(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
			return
		}
		logger.Error("Failed to fetch profile for user %d: %v", userID, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch profile", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, stats)
}
