package routes

import (
	"poc-gin/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configure toutes les routes de l'application
func SetupAuthRoutes(r *gin.Engine, authHandler *controllers.AuthHandler) {

	r.POST("/login", authHandler.Login)
	r.POST("/register", authHandler.Register)

}
