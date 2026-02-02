package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(r *gin.Engine, productHandler *controllers.ProductHandler) {

	protected := r.Group("/products")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("", productHandler.FindProduct)
		protected.GET("/:id", productHandler.FindOneProduct)
		protected.POST("", productHandler.CreateProduct)
		protected.PUT("/:id", productHandler.UpdateProduct)
		protected.DELETE("/:id", productHandler.DeleteProduct)
	}
}
