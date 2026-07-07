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

	sellerProducts := r.Group("/seller/products")
	sellerProducts.Use(authMW.Authenticate())
	{
		sellerProducts.GET("", productHandler.FindSellerProducts)
	}

	authProducts := r.Group("/products")
	authProducts.Use(authMW.Authenticate())
	{
		authProducts.POST("", productHandler.CreateProduct)
		authProducts.PUT("/:id", productHandler.UpdateProduct)
		authProducts.DELETE("/:id", productHandler.DeleteProduct)
		authProducts.POST("/:id/image", productHandler.UploadProductImage)
		authProducts.DELETE("/:id/image", productHandler.DeleteProductImage)
	}
}
