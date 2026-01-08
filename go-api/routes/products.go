package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configure toutes les routes de l'application
func SetupProductRoutes(r *gin.Engine, productHandler *controllers.ProductHandler) {

	r.GET("/products", productHandler.FindProduct)
	r.GET("/products/:id", productHandler.FindOneProduct)
	protected := r.Group("/products")
	protected.Use(middlewares.AuthMiddleware())
	{
		r.POST("/products", productHandler.CreateProduct)

		r.PUT("/products/:id", productHandler.UpdateProduct)

		r.DELETE("/products/:id", productHandler.DeleteProduct)
	}

}
