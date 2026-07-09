package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(r *gin.Engine, profileHandler *controllers.ProfileHandler, authMW *middlewares.AuthMiddleware) {
	profile := r.Group("/profile")
	profile.Use(authMW.Authenticate())
	{
		profile.GET("", profileHandler.GetProfile)
	}
}
