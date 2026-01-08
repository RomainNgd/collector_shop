package main

import (
	"log"
	"poc-gin/controllers"
	"poc-gin/database"
	"poc-gin/models"
	"poc-gin/routes"
	"poc-gin/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	database.ConnectDB()

	database.DB.AutoMigrate(&models.Product{})
	database.DB.AutoMigrate(&models.User{})

	r := gin.Default()

	productHandler := &controllers.ProductHandler{
		Service: services.NewProductService(),
	}

	authHandler := &controllers.AuthHandler{
		Service: services.NewAuthService(),
	}

	routes.SetupProductRoutes(r, productHandler)
	routes.SetupAuthRoutes(r, authHandler)

	r.Run(":8080")
}
