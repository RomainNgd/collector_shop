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

func TestCategoryHandlerFindCategory(t *testing.T) {
	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			getAllFn: func() ([]*models.Category, error) {
				return []*models.Category{{Name: "Cards"}}, nil
			},
		})

		recorder := performJSONRequest(handler.FindCategory, http.MethodGet, "/categories", nil)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})

	t.Run("returns 500 on error", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			getAllFn: func() ([]*models.Category, error) { return nil, errors.New("db error") },
		})

		recorder := performJSONRequest(handler.FindCategory, http.MethodGet, "/categories", nil)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})
}

func TestCategoryHandlerFindOneCategory(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{})
		recorder := performJSONRequest(handler.FindOneCategory, http.MethodGet, "/categories/abc", nil, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return nil, gorm.ErrRecordNotFound },
		})
		recorder := performJSONRequest(handler.FindOneCategory, http.MethodGet, "/categories/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 when found", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{Name: "Cards"}, nil },
		})
		recorder := performJSONRequest(handler.FindOneCategory, http.MethodGet, "/categories/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestCategoryHandlerCreateCategory(t *testing.T) {
	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{})
		recorder := performJSONRequest(handler.CreateCategory, http.MethodPost, "/categories", map[string]any{"name": "A"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 500 on service error", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			createFn: func(category *models.Category) error { return errors.New("db error") },
		})
		recorder := performJSONRequest(handler.CreateCategory, http.MethodPost, "/categories", map[string]any{
			"name":        "Cards",
			"description": "Trading cards",
		})
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("returns 201 on success", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			createFn: func(category *models.Category) error { return nil },
		})
		recorder := performJSONRequest(handler.CreateCategory, http.MethodPost, "/categories", map[string]any{
			"name":        "Cards",
			"description": "Trading cards",
		})
		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", recorder.Code)
		}
	})
}

func TestCategoryHandlerUpdateCategory(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{})
		recorder := performJSONRequest(handler.UpdateCategory, http.MethodPut, "/categories/abc", map[string]any{
			"name":        "Cards",
			"description": "Trading cards",
		}, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{})
		recorder := performJSONRequest(handler.UpdateCategory, http.MethodPut, "/categories/1", map[string]any{
			"name": "A",
		}, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			updateFn: func(id uint, updates map[string]interface{}) (*models.Category, error) {
				return nil, gorm.ErrRecordNotFound
			},
		})
		recorder := performJSONRequest(handler.UpdateCategory, http.MethodPut, "/categories/1", map[string]any{
			"name":        "Cards",
			"description": "Trading cards",
		}, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			updateFn: func(id uint, updates map[string]interface{}) (*models.Category, error) {
				return &models.Category{Name: "Cards"}, nil
			},
		})
		recorder := performJSONRequest(handler.UpdateCategory, http.MethodPut, "/categories/1", map[string]any{
			"name":        "Cards",
			"description": "Trading cards",
		}, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestCategoryHandlerDeleteCategory(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{})
		recorder := performJSONRequest(handler.DeleteCategory, http.MethodDelete, "/categories/abc", nil, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when not found", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			deleteFn: func(id uint) error { return gorm.ErrRecordNotFound },
		})
		recorder := performJSONRequest(handler.DeleteCategory, http.MethodDelete, "/categories/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 409 when category is in use", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			deleteFn: func(id uint) error { return services.ErrCategoryInUse },
		})
		recorder := performJSONRequest(handler.DeleteCategory, http.MethodDelete, "/categories/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewCategoryHandler(&mockCategoryService{
			deleteFn: func(id uint) error { return nil },
		})
		recorder := performJSONRequest(handler.DeleteCategory, http.MethodDelete, "/categories/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}
