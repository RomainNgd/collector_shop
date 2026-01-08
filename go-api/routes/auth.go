package routes

import (
	"poc-gin/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configure toutes les routes de l'application
func SetupAuthRoutes(r *gin.Engine, authHandler *controllers.AuthHandler) {

	r.GET("/login", authHandler.Login)
	r.GET("/register", authHandler.Register)

}
