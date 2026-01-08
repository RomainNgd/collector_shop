package routes

import (
	"poc-gin/controllers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configure toutes les routes de l'application
func SetupRoutes(r *gin.Engine, h *controllers.Handler) {

	r.GET("/products", h.FindProduct)

	r.GET("/products/:id", h.FindOneProduct)

	r.POST("/products", h.CreateProduct)

	r.PUT("/products/:id", h.UpdateProduct)

	r.DELETE("/products/:id", h.DeleteProduct)

}
