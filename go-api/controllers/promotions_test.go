package controllers

import (
	"errors"
	"net/http"
	"poc-gin/models"
	"poc-gin/services"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func promotionPayload() map[string]any {
	return map[string]any{
		"name":           "Printemps",
		"description":    "Promo globale",
		"type":           models.PromotionTypePercentage,
		"value":          10,
		"is_active":      true,
		"applies_to_all": false,
		"product_ids":    []uint{1, 2},
	}
}

func TestPromotionHandlerFindPromotion(t *testing.T) {
	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{
			getAllFn: func() ([]*models.Promotion, error) {
				return []*models.Promotion{{Name: "Printemps"}}, nil
			},
		})

		recorder := performJSONRequest(handler.FindPromotion, http.MethodGet, "/promotions", nil)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})

	t.Run("returns 500 on error", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{
			getAllFn: func() ([]*models.Promotion, error) { return nil, errors.New("db error") },
		})

		recorder := performJSONRequest(handler.FindPromotion, http.MethodGet, "/promotions", nil)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})
}

func TestPromotionHandlerFindOnePromotion(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{})
		recorder := performJSONRequest(handler.FindOnePromotion, http.MethodGet, "/promotions/abc", nil, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when promotion is missing", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{
			getByIDFn: func(id uint) (*models.Promotion, error) { return nil, gorm.ErrRecordNotFound },
		})
		recorder := performJSONRequest(handler.FindOnePromotion, http.MethodGet, "/promotions/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 when promotion exists", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{
			getByIDFn: func(id uint) (*models.Promotion, error) { return &models.Promotion{Name: "Printemps"}, nil },
		})
		recorder := performJSONRequest(handler.FindOnePromotion, http.MethodGet, "/promotions/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestPromotionHandlerCreatePromotion(t *testing.T) {
	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{})
		recorder := performJSONRequest(handler.CreatePromotion, http.MethodPost, "/promotions", map[string]any{"name": "A"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("maps validation errors to 400", func(t *testing.T) {
		testCases := []error{
			services.ErrInvalidPromotionType,
			services.ErrInvalidPromotionValue,
			services.ErrPromotionProductsEmpty,
			services.ErrPromotionProductsNotFound,
		}

		for _, expectedErr := range testCases {
			handler := NewPromotionHandler(&mockPromotionService{
				createFn: func(input services.PromotionInput) (*models.Promotion, error) {
					return nil, expectedErr
				},
			})
			recorder := performJSONRequest(handler.CreatePromotion, http.MethodPost, "/promotions", promotionPayload())
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %v, got %d", expectedErr, recorder.Code)
			}
		}
	})

	t.Run("returns 201 on success", func(t *testing.T) {
		var captured services.PromotionInput
		handler := NewPromotionHandler(&mockPromotionService{
			createFn: func(input services.PromotionInput) (*models.Promotion, error) {
				captured = input
				return &models.Promotion{Name: input.Name}, nil
			},
		})

		recorder := performJSONRequest(handler.CreatePromotion, http.MethodPost, "/promotions", promotionPayload())
		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", recorder.Code)
		}
		if len(captured.ProductIDs) != 2 || !captured.IsActive || captured.AppliesToAll {
			t.Fatalf("unexpected captured input: %#v", captured)
		}
	})
}

func TestPromotionHandlerUpdatePromotion(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{})
		recorder := performJSONRequest(handler.UpdatePromotion, http.MethodPut, "/promotions/abc", promotionPayload(), gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when promotion is missing", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{
			updateFn: func(id uint, input services.PromotionInput) (*models.Promotion, error) {
				return nil, gorm.ErrRecordNotFound
			},
		})
		recorder := performJSONRequest(handler.UpdatePromotion, http.MethodPut, "/promotions/1", promotionPayload(), gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		var capturedID uint
		var captured services.PromotionInput
		handler := NewPromotionHandler(&mockPromotionService{
			updateFn: func(id uint, input services.PromotionInput) (*models.Promotion, error) {
				capturedID = id
				captured = input
				return &models.Promotion{Name: input.Name}, nil
			},
		})
		recorder := performJSONRequest(handler.UpdatePromotion, http.MethodPut, "/promotions/1", promotionPayload(), gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		if capturedID != 1 || captured.Type != models.PromotionTypePercentage {
			t.Fatalf("unexpected captured update: id=%d input=%#v", capturedID, captured)
		}
	})
}

func TestPromotionHandlerDeletePromotion(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{})
		recorder := performJSONRequest(handler.DeletePromotion, http.MethodDelete, "/promotions/abc", nil, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when promotion is missing", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{
			deleteFn: func(id uint) error { return gorm.ErrRecordNotFound },
		})
		recorder := performJSONRequest(handler.DeletePromotion, http.MethodDelete, "/promotions/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewPromotionHandler(&mockPromotionService{
			deleteFn: func(id uint) error { return nil },
		})
		recorder := performJSONRequest(handler.DeletePromotion, http.MethodDelete, "/promotions/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}
