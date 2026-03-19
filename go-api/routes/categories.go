package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupCategoryRoutes(r *gin.Engine, categoryHandler *controllers.CategoryHandler, authMW *middlewares.AuthMiddleware) {
	categories := r.Group("/categories")
	{
		categories.GET("", categoryHandler.FindCategory)
		categories.GET("/:id", categoryHandler.FindOneCategory)
	}

	adminCategories := r.Group("/categories")
	adminCategories.Use(authMW.Authenticate(), authMW.RequireAdmin())
	{
		adminCategories.POST("", categoryHandler.CreateCategory)
		adminCategories.PUT("/:id", categoryHandler.UpdateCategory)
		adminCategories.DELETE("/:id", categoryHandler.DeleteCategory)
	}
}
