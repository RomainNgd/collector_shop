package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(r *gin.Engine, productHandler *controllers.ProductHandler, authMW *middlewares.AuthMiddleware) {
	products := r.Group("/products")
	{
		products.GET("", productHandler.FindProduct)
		products.GET("/:id", productHandler.FindOneProduct)
	}

	adminProducts := r.Group("/products")
	adminProducts.Use(authMW.Authenticate(), authMW.RequireAdmin())
	{
		adminProducts.POST("", productHandler.CreateProduct)
		adminProducts.PUT("/:id", productHandler.UpdateProduct)
		adminProducts.DELETE("/:id", productHandler.DeleteProduct)
		adminProducts.POST("/:id/image", productHandler.UploadProductImage)
		adminProducts.DELETE("/:id/image", productHandler.DeleteProductImage)
	}
}
