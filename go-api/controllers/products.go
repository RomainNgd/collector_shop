package controllers

import (
	"log"
	"net/http"
	"poc-gin/services"

	"github.com/gin-gonic/gin"
)
import "poc-gin/models"

type ProductHandler struct {
	Service *services.ProductService
}

func (h *ProductHandler) FindProduct(c *gin.Context) {
	products, err := h.Service.GetAllProducts()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error with products list"})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) FindOneProduct(c *gin.Context) {
	id := c.Param("id")
	product, err := h.Service.GetProduct(id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error with product"})
		return
	}
	c.JSON(200, product)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	product, err := h.Service.CreateProduct(input)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error with product"})
		return
	}
	c.JSON(201, product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	product, err := h.Service.GetProduct(id)
	if err != nil || id == "0" {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	product, err = h.Service.UpdateProduct(product, input)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error with product"})
		return
	}
	c.JSON(200, product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	err := h.Service.DeleteProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error with product"})
		return
	}
	c.JSON(200, gin.H{"message": "deleted"})
}
