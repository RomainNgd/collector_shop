package routes

import (
	"poc-gin/controllers"

	"github.com/gin-gonic/gin"
)

func SetupHealthRoutes(r *gin.Engine, healthHandler *controllers.HealthHandler) {
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)
}
