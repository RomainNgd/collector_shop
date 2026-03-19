package controllers

import (
	"errors"
	"net/http"
	"poc-gin/models"
	"poc-gin/pkg/logger"
	"poc-gin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	categoryService services.CategoryServiceInterface
}

func NewCategoryHandler(categoryService services.CategoryServiceInterface) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) FindCategory(c *gin.Context) {
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		logger.Error("Failed to fetch categories: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch categories", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, categories)
}

func (h *CategoryHandler) FindOneCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID", nil)
		return
	}

	category, err := h.categoryService.GetCategoryByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found", nil)
			return
		}
		logger.Error("Failed to fetch category %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch category", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, category)
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", err.Error())
		return
	}

	category := &models.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.categoryService.CreateCategory(category); err != nil {
		logger.Error("Failed to create category: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create category", nil)
		return
	}

	RespondSuccess(c, http.StatusCreated, category)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID", nil)
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", err.Error())
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
	}

	category, err := h.categoryService.UpdateCategory(uint(id), updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found", nil)
			return
		}
		logger.Error("Failed to update category %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, category)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID", nil)
		return
	}

	if err := h.categoryService.DeleteCategory(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found", nil)
			return
		}
		if errors.Is(err, services.ErrCategoryInUse) {
			RespondError(c, http.StatusConflict, "CATEGORY_IN_USE", "Category is linked to existing products", nil)
			return
		}
		logger.Error("Failed to delete category %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
