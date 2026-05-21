package routes

import (
	"poc-gin/controllers"

	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(r *gin.Engine, paymentHandler *controllers.PaymentHandler) {
	payments := r.Group("/payments")
	{
		payments.POST("/stripe/webhook", paymentHandler.HandleStripeWebhook)
	}
}
