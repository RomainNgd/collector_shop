package controllers

import (
	"errors"
	"net/http"
	"poc-gin/pkg/logger"
	"poc-gin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PromotionHandler struct {
	promotionService services.PromotionServiceInterface
}

func NewPromotionHandler(promotionService services.PromotionServiceInterface) *PromotionHandler {
	return &PromotionHandler{
		promotionService: promotionService,
	}
}

func (h *PromotionHandler) FindPromotion(c *gin.Context) {
	ctx := c.Request.Context()

	promotions, err := h.promotionService.GetAllPromotions(ctx)
	if err != nil {
		logger.Error("Failed to fetch promotions: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch promotions", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, promotions)
}

func (h *PromotionHandler) FindOnePromotion(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidPromotionID, nil)
		return
	}

	promotion, err := h.promotionService.GetPromotionByID(ctx, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PROMOTION_NOT_FOUND", errorPromotionNotFound, nil)
			return
		}
		logger.Error("Failed to fetch promotion %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch promotion", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, promotion)
}

func (h *PromotionHandler) CreatePromotion(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	promotion, err := h.promotionService.CreatePromotion(ctx, services.PromotionInput{
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Value:        req.Value,
		IsActive:     dereferenceBool(req.IsActive),
		AppliesToAll: dereferenceBool(req.AppliesToAll),
		ProductIDs:   req.ProductIDs,
	})
	if err != nil {
		if handlePromotionError(c, err) {
			return
		}
		logger.Error("Failed to create promotion: %v", err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create promotion", nil)
		return
	}

	RespondSuccess(c, http.StatusCreated, promotion)
}

func (h *PromotionHandler) UpdatePromotion(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidPromotionID, nil)
		return
	}

	var req UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", errorInvalidRequestPayload, err.Error())
		return
	}

	promotion, err := h.promotionService.UpdatePromotion(ctx, uint(id), services.PromotionInput{
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Value:        req.Value,
		IsActive:     dereferenceBool(req.IsActive),
		AppliesToAll: dereferenceBool(req.AppliesToAll),
		ProductIDs:   req.ProductIDs,
	})
	if err != nil {
		if handlePromotionError(c, err) {
			return
		}
		logger.Error("Failed to update promotion %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update promotion", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, promotion)
}

func (h *PromotionHandler) DeletePromotion(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", errorInvalidPromotionID, nil)
		return
	}

	if err := h.promotionService.DeletePromotion(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "PROMOTION_NOT_FOUND", errorPromotionNotFound, nil)
			return
		}
		logger.Error("Failed to delete promotion %d: %v", id, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete promotion", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, gin.H{"message": "Promotion deleted successfully"})
}

func dereferenceBool(value *bool) bool {
	return value != nil && *value
}

func handlePromotionError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		RespondError(c, http.StatusNotFound, "PROMOTION_NOT_FOUND", errorPromotionNotFound, nil)
		return true
	case errors.Is(err, services.ErrInvalidPromotionType):
		RespondError(c, http.StatusBadRequest, "INVALID_PROMOTION_TYPE", "Promotion type is invalid", nil)
		return true
	case errors.Is(err, services.ErrInvalidPromotionValue):
		RespondError(c, http.StatusBadRequest, "INVALID_PROMOTION_VALUE", "Promotion value is invalid", nil)
		return true
	case errors.Is(err, services.ErrPromotionProductsEmpty):
		RespondError(c, http.StatusBadRequest, "PROMOTION_PRODUCTS_REQUIRED", "Select at least one product or enable the global scope", nil)
		return true
	case errors.Is(err, services.ErrPromotionProductsNotFound):
		RespondError(c, http.StatusBadRequest, "PROMOTION_PRODUCTS_NOT_FOUND", "Some selected products do not exist", nil)
		return true
	default:
		return false
	}
}
