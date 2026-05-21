package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupPromotionRoutes(r *gin.Engine, promotionHandler *controllers.PromotionHandler, authMW *middlewares.AuthMiddleware) {
	promotions := r.Group("/promotions")
	promotions.Use(authMW.Authenticate(), authMW.RequireAdmin())
	{
		promotions.GET("", promotionHandler.FindPromotion)
		promotions.GET("/:id", promotionHandler.FindOnePromotion)
		promotions.POST("", promotionHandler.CreatePromotion)
		promotions.PUT("/:id", promotionHandler.UpdatePromotion)
		promotions.DELETE("/:id", promotionHandler.DeletePromotion)
	}
}
