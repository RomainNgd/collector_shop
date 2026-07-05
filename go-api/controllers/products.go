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

type ProductHandler struct {
	productService  services.ProductServiceInterface
	categoryService services.CategoryServiceInterface
	fileService     services.FileServiceInterface
}

func NewProductHandler(productService services.ProductServiceInterface, categoryService services.CategoryServiceInterface, fileService services.FileServiceInterface) *ProductHandler {
	return &ProductHandler{
		productService:  productService,
		categoryService: categoryService,
		fileService:     fileService,
	}
}

func (h *ProductHandler) FindProduct(c *gin.Context) {
	ctx := c.Request.Context()

	products, err := h.productService.GetAllProducts(ctx)
	if err != nil {
		logger.Error("Failed to fetch products: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch products", nil)
		return
	}
	RespondSuccess(c, http.StatusOK, products)
}

func (h *ProductHandler) FindOneProduct(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	product, err := h.productService.GetProductByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		logger.Error(logFailedFetchProduct, id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedFetchProduct, nil)
		return
	}

	RespondSuccess(c, http.StatusOK, product)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
	}

	if _, err := h.categoryService.GetCategoryByID(ctx, req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusBadRequest, "CATEGORY_NOT_FOUND", errorCategoryNotFound, nil)
			return
		}
		logger.Error("Failed to fetch category %d: %v", req.CategoryID, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to validate category", nil)
		return
	}

	if err := h.productService.CreateProduct(ctx, product); err != nil {
		logger.Error("Failed to create product: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create product", nil)
		return
	}

	RespondSuccess(c, http.StatusCreated, product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	updates := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"image":       req.Image,
		"price":       req.Price,
		"category_id": req.CategoryID,
	}

	if _, err := h.categoryService.GetCategoryByID(ctx, req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusBadRequest, "CATEGORY_NOT_FOUND", errorCategoryNotFound, nil)
			return
		}
		logger.Error("Failed to fetch category %d: %v", req.CategoryID, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to validate category", nil)
		return
	}

	product, err := h.productService.UpdateProduct(ctx, uint(id), updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		logger.Error("Failed to update product %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedUpdateProduct, nil)
		return
	}

	RespondSuccess(c, http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	if err := h.productService.DeleteProduct(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		logger.Error("Failed to delete product %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete product", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func (h *ProductHandler) UploadProductImage(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	product, err := h.productService.GetProductByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		logger.Error(logFailedFetchProduct, id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedFetchProduct, nil)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		RespondError(c, http.StatusBadRequest, "FILE_REQUIRED", "Image file is required", err.Error())
		return
	}

	filename, err := h.fileService.SaveImage(file)
	if err != nil {
		if errors.Is(err, services.ErrFileTooLarge) {
			RespondError(c, http.StatusBadRequest, "FILE_TOO_LARGE", "File size exceeds maximum allowed", nil)
			return
		}
		if errors.Is(err, services.ErrInvalidFileFormat) {
			RespondError(c, http.StatusBadRequest, "INVALID_FILE_FORMAT", "Invalid file format", nil)
			return
		}
		logger.Error("Failed to save image: %v", err)
		RespondError(c, http.StatusInternalServerError, "FILE_UPLOAD_ERROR", "Failed to upload image", nil)
		return
	}

	if product.Image != "" {
		_ = h.fileService.DeleteImage(product.Image)
	}

	updated, err := h.productService.UpdateProduct(ctx, uint(id), map[string]interface{}{"image": filename})
	if err != nil {
		logger.Error("Failed to update product %d with image: %v", id, err)
		_ = h.fileService.DeleteImage(filename)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedUpdateProduct, nil)
		return
	}

	RespondSuccess(c, http.StatusOK, updated)
}

func (h *ProductHandler) DeleteProductImage(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	product, err := h.productService.GetProductByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		logger.Error(logFailedFetchProduct, id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedFetchProduct, nil)
		return
	}

	if product.Image != "" {
		if err := h.fileService.DeleteImage(product.Image); err != nil {
			logger.Warn("Failed to delete image file: %v", err)
		}
	}

	updated, err := h.productService.UpdateProduct(ctx, uint(id), map[string]interface{}{"image": ""})
	if err != nil {
		logger.Error("Failed to update product %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedUpdateProduct, nil)
		return
	}

	RespondSuccess(c, http.StatusOK, updated)
}
