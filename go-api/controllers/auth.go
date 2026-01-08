package controllers

import (
	"net/http"
	"poc-gin/models"
	"poc-gin/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Service *services.AuthService
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Service.Register(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de créer l'utilisateur"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"email": user.Email, "id": user.ID})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input models.User // On réutilise le struct User juste pour binder email/pass
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.Service.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email ou mot de passe invalide"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
