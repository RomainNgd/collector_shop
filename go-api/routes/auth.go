package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine, authHandler *controllers.AuthHandler, rateLimiter *middlewares.RateLimiter) {
	auth := r.Group("/auth")
	if rateLimiter != nil {
		auth.Use(rateLimiter.Middleware())
	}
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
	}
}
