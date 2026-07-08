package controllers

import (
	"errors"
	"net/http"
	"poc-gin/models"
	"poc-gin/pkg/constants"
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

	var excludeSellerID *uint
	if userID, err := userIDFromContext(c); err == nil {
		excludeSellerID = &userID
	}

	limit, offset := paginationParams(c)
	products, total, err := h.productService.GetAllProducts(ctx, excludeSellerID, services.Pagination{Limit: limit, Offset: offset})
	if err != nil {
		logger.Error("Failed to fetch products: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch products", nil)
		return
	}
	setPaginationHeaders(c, total, limit, offset)
	RespondSuccess(c, http.StatusOK, products)
}

func (h *ProductHandler) FindSellerProducts(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", errorInvalidAuthenticationContext, nil)
		return
	}
	if userRoleFromContext(c) == constants.RoleAdmin {
		RespondError(c, http.StatusForbidden, "PRODUCT_ADMIN_CREATE_FORBIDDEN", "Admins cannot sell products", nil)
		return
	}

	products, err := h.productService.GetProductsForSeller(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch seller products for user %d: %v", userID, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch seller products", nil)
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

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", errorInvalidAuthenticationContext, nil)
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	promotionActive := dereferenceBool(req.PromotionActive)
	product := &models.Product{
		Name:            req.Name,
		Description:     req.Description,
		Image:           req.Image,
		Price:           req.Price,
		Stock:           req.Stock,
		IsActive:        true,
		SellerID:        &userID,
		CategoryID:      req.CategoryID,
		PromotionType:   req.PromotionType,
		PromotionValue:  req.PromotionValue,
		PromotionActive: promotionActive,
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
		if handleProductError(c, err) {
			return
		}
		logger.Error("Failed to create product: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create product", nil)
		return
	}

	RespondSuccess(c, http.StatusCreated, product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", errorInvalidAuthenticationContext, nil)
		return
	}
	role := userRoleFromContext(c)

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
		"name":             req.Name,
		"description":      req.Description,
		"image":            req.Image,
		"price":            req.Price,
		"stock":            req.Stock,
		"is_active":        dereferenceBool(req.IsActive),
		"category_id":      req.CategoryID,
		"promotion_type":   req.PromotionType,
		"promotion_value":  req.PromotionValue,
		"promotion_active": dereferenceBool(req.PromotionActive),
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

	product, err := h.productService.UpdateProduct(ctx, userID, role, uint(id), updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		if handleProductError(c, err) {
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

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", errorInvalidAuthenticationContext, nil)
		return
	}
	role := userRoleFromContext(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	if err := h.productService.DeleteProduct(ctx, userID, role, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		if handleProductError(c, err) {
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

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", errorInvalidAuthenticationContext, nil)
		return
	}
	role := userRoleFromContext(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	product, err := h.productService.GetProductForManagement(ctx, userID, role, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		if handleProductError(c, err) {
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

	updated, err := h.productService.UpdateProduct(ctx, userID, role, uint(id), map[string]interface{}{"image": filename})
	if err != nil {
		if handleProductError(c, err) {
			_ = h.fileService.DeleteImage(filename)
			return
		}
		logger.Error("Failed to update product %d with image: %v", id, err)
		_ = h.fileService.DeleteImage(filename)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedUpdateProduct, nil)
		return
	}

	// Remove the previous file only once the new image is persisted, so a
	// failed update never destroys the image still referenced in database.
	if product.Image != "" && product.Image != filename {
		_ = h.fileService.DeleteImage(product.Image)
	}

	RespondSuccess(c, http.StatusOK, updated)
}

func (h *ProductHandler) DeleteProductImage(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", errorInvalidAuthenticationContext, nil)
		return
	}
	role := userRoleFromContext(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidProductID, nil)
		return
	}

	product, err := h.productService.GetProductForManagement(ctx, userID, role, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", errorProductNotFound, nil)
			return
		}
		if handleProductError(c, err) {
			return
		}
		logger.Error(logFailedFetchProduct, id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedFetchProduct, nil)
		return
	}

	updated, err := h.productService.UpdateProduct(ctx, userID, role, uint(id), map[string]interface{}{"image": ""})
	if err != nil {
		if handleProductError(c, err) {
			return
		}
		logger.Error("Failed to update product %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", errorFailedUpdateProduct, nil)
		return
	}

	// Remove the file only after the database no longer references it.
	if product.Image != "" {
		if err := h.fileService.DeleteImage(product.Image); err != nil {
			logger.Warn("Failed to delete image file: %v", err)
		}
	}

	RespondSuccess(c, http.StatusOK, updated)
}

func handleProductError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, services.ErrProductAccessDenied):
		RespondError(c, http.StatusForbidden, "PRODUCT_ACCESS_DENIED", "You cannot modify this product", nil)
		return true
	case errors.Is(err, services.ErrProductSellerRequired):
		RespondError(c, http.StatusBadRequest, "PRODUCT_SELLER_REQUIRED", "Product seller is required", nil)
		return true
	case errors.Is(err, services.ErrProductInvalidStock):
		RespondError(c, http.StatusBadRequest, "PRODUCT_STOCK_INVALID", "Product stock must be positive", nil)
		return true
	case errors.Is(err, services.ErrProductInvalidPromotionType):
		RespondError(c, http.StatusBadRequest, "PRODUCT_PROMOTION_TYPE_INVALID", "Promotion type is invalid", nil)
		return true
	case errors.Is(err, services.ErrProductInvalidPromotionValue):
		RespondError(c, http.StatusBadRequest, "PRODUCT_PROMOTION_VALUE_INVALID", "Promotion value is invalid", nil)
		return true
	default:
		return false
	}
}
