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

	h := &controllers.Handler{
		Service: services.NewProductService(),
	}
	r := gin.Default()

	routes.SetupRoutes(r, h)

	r.Run(":8080")
}
