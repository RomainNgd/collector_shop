package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupOrderRoutes(r *gin.Engine, orderHandler *controllers.OrderHandler, authMW *middlewares.AuthMiddleware) {
	orders := r.Group("/orders")
	orders.Use(authMW.Authenticate())
	{
		orders.GET("", orderHandler.FindOrder)
		orders.GET("/:id", orderHandler.FindOneOrder)
		orders.POST("", orderHandler.CreateOrder)
		orders.POST("/:id/checkout-session", orderHandler.CreateCheckoutSession)
		orders.PUT("/:id", orderHandler.UpdateOrder)
		orders.DELETE("/:id", orderHandler.DeleteOrder)
	}

	sellerStats := r.Group("/seller/stats")
	sellerStats.Use(authMW.Authenticate())
	{
		sellerStats.GET("", orderHandler.FindSellerStats)
	}
}
