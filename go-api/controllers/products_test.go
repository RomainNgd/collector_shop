package controllers

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"poc-gin/models"
	"poc-gin/services"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func productPayload() map[string]any {
	return map[string]any{
		"name":        "Blue Eyes",
		"description": "Mint card",
		"image":       "image.png",
		"price":       19.99,
		"category_id": 2,
	}
}

func TestProductHandlerFindProduct(t *testing.T) {
	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getAllFn: func() ([]*models.Product, error) {
				return []*models.Product{{Name: "Blue Eyes"}}, nil
			},
		}, &mockCategoryService{}, &mockFileService{})

		recorder := performJSONRequest(handler.FindProduct, http.MethodGet, "/products", nil)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})

	t.Run("returns 500 on error", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getAllFn: func() ([]*models.Product, error) { return nil, errors.New("db error") },
		}, &mockCategoryService{}, &mockFileService{})

		recorder := performJSONRequest(handler.FindProduct, http.MethodGet, "/products", nil)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})
}

func TestProductHandlerFindOneProduct(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.FindOneProduct, http.MethodGet, "/products/abc", nil, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when product is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return nil, gorm.ErrRecordNotFound },
		}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.FindOneProduct, http.MethodGet, "/products/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 when product exists", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return &models.Product{Name: "Blue Eyes"}, nil },
		}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.FindOneProduct, http.MethodGet, "/products/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestProductHandlerCreateProduct(t *testing.T) {
	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.CreateProduct, http.MethodPost, "/products", map[string]any{"name": "A"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 400 when category is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return nil, gorm.ErrRecordNotFound },
		}, &mockFileService{})
		recorder := performJSONRequest(handler.CreateProduct, http.MethodPost, "/products", productPayload())
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 500 when category lookup fails", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return nil, errors.New("db error") },
		}, &mockFileService{})
		recorder := performJSONRequest(handler.CreateProduct, http.MethodPost, "/products", productPayload())
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("returns 201 on success", func(t *testing.T) {
		var created *models.Product
		handler := NewProductHandler(&mockProductService{
			createFn: func(product *models.Product) error {
				created = product
				return nil
			},
		}, &mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{}, nil },
		}, &mockFileService{})

		recorder := performJSONRequest(handler.CreateProduct, http.MethodPost, "/products", productPayload())
		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", recorder.Code)
		}
		if created == nil || created.CategoryID != 2 {
			t.Fatalf("expected category id 2, got %#v", created)
		}
	})
}

func TestProductHandlerUpdateProduct(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.UpdateProduct, http.MethodPut, "/products/abc", productPayload(), gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.UpdateProduct, http.MethodPut, "/products/1", map[string]any{"name": "A"}, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 400 when category is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return nil, gorm.ErrRecordNotFound },
		}, &mockFileService{})
		recorder := performJSONRequest(handler.UpdateProduct, http.MethodPut, "/products/1", productPayload(), gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when product is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			updateFn: func(id uint, updates map[string]interface{}) (*models.Product, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}, &mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{}, nil },
		}, &mockFileService{})
		recorder := performJSONRequest(handler.UpdateProduct, http.MethodPut, "/products/1", productPayload(), gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		var updates map[string]interface{}
		handler := NewProductHandler(&mockProductService{
			updateFn: func(id uint, in map[string]interface{}) (*models.Product, error) {
				updates = in
				return &models.Product{Name: "Blue Eyes", CategoryID: 2}, nil
			},
		}, &mockCategoryService{
			getByIDFn: func(id uint) (*models.Category, error) { return &models.Category{}, nil },
		}, &mockFileService{})
		recorder := performJSONRequest(handler.UpdateProduct, http.MethodPut, "/products/1", productPayload(), gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		if updates["category_id"] != uint(2) {
			t.Fatalf("expected category update to 2, got %#v", updates["category_id"])
		}
	})
}

func TestProductHandlerDeleteProduct(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.DeleteProduct, http.MethodDelete, "/products/abc", nil, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when product is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			deleteFn: func(id uint) error { return gorm.ErrRecordNotFound },
		}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.DeleteProduct, http.MethodDelete, "/products/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			deleteFn: func(id uint) error { return nil },
		}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.DeleteProduct, http.MethodDelete, "/products/1", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func newMultipartImageRequest(t *testing.T, method, target, fieldName, filename string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create multipart: %v", err)
	}
	if _, err := part.Write([]byte("fake-image-content")); err != nil {
		t.Fatalf("failed to write multipart: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequest(method, target, &body)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func performMultipartRequest(handlerFunc gin.HandlerFunc, req *http.Request, params ...gin.Param) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req
	ctx.Params = params
	handlerFunc(ctx)
	return recorder
}

func TestProductHandlerUploadProductImage(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{}, &mockFileService{})
		req, _ := http.NewRequest(http.MethodPost, "/products/abc/image", strings.NewReader(""))
		recorder := performMultipartRequest(handler.UploadProductImage, req, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when product is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return nil, gorm.ErrRecordNotFound },
		}, &mockCategoryService{}, &mockFileService{})
		req, _ := http.NewRequest(http.MethodPost, "/products/1/image", strings.NewReader(""))
		recorder := performMultipartRequest(handler.UploadProductImage, req, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 400 when file is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return &models.Product{}, nil },
		}, &mockCategoryService{}, &mockFileService{})
		req, _ := http.NewRequest(http.MethodPost, "/products/1/image", strings.NewReader(""))
		recorder := performMultipartRequest(handler.UploadProductImage, req, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("maps file size and format errors to 400", func(t *testing.T) {
		testCases := []struct {
			name string
			err  error
		}{
			{name: "too large", err: services.ErrFileTooLarge},
			{name: "invalid format", err: services.ErrInvalidFileFormat},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				handler := NewProductHandler(&mockProductService{
					getByIDFn: func(id uint) (*models.Product, error) { return &models.Product{}, nil },
				}, &mockCategoryService{}, &mockFileService{
					saveFn: func(file *multipart.FileHeader) (string, error) { return "", tc.err },
				})
				req := newMultipartImageRequest(t, http.MethodPost, "/products/1/image", "image", "photo.png")
				recorder := performMultipartRequest(handler.UploadProductImage, req, gin.Param{Key: "id", Value: "1"})
				if recorder.Code != http.StatusBadRequest {
					t.Fatalf("expected 400, got %d", recorder.Code)
				}
			})
		}
	})

	t.Run("returns 500 when save image fails", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return &models.Product{}, nil },
		}, &mockCategoryService{}, &mockFileService{
			saveFn: func(file *multipart.FileHeader) (string, error) { return "", errors.New("disk full") },
		})
		req := newMultipartImageRequest(t, http.MethodPost, "/products/1/image", "image", "photo.png")
		recorder := performMultipartRequest(handler.UploadProductImage, req, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("deletes previous image and returns 200 on success", func(t *testing.T) {
		var deletedOld string
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) {
				return &models.Product{Image: "old.png"}, nil
			},
			updateFn: func(id uint, updates map[string]interface{}) (*models.Product, error) {
				return &models.Product{Image: updates["image"].(string)}, nil
			},
		}, &mockCategoryService{}, &mockFileService{
			saveFn: func(file *multipart.FileHeader) (string, error) { return "new.png", nil },
			deleteFn: func(filename string) error {
				deletedOld = filename
				return nil
			},
		})
		req := newMultipartImageRequest(t, http.MethodPost, "/products/1/image", "image", "photo.png")
		recorder := performMultipartRequest(handler.UploadProductImage, req, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		if deletedOld != "old.png" {
			t.Fatalf("expected old image deletion, got %q", deletedOld)
		}
	})

	t.Run("deletes newly uploaded image when product update fails", func(t *testing.T) {
		var deletedNew string
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return &models.Product{}, nil },
			updateFn: func(id uint, updates map[string]interface{}) (*models.Product, error) {
				return nil, errors.New("db error")
			},
		}, &mockCategoryService{}, &mockFileService{
			saveFn: func(file *multipart.FileHeader) (string, error) { return "new.png", nil },
			deleteFn: func(filename string) error {
				deletedNew = filename
				return nil
			},
		})
		req := newMultipartImageRequest(t, http.MethodPost, "/products/1/image", "image", "photo.png")
		recorder := performMultipartRequest(handler.UploadProductImage, req, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
		if deletedNew != "new.png" {
			t.Fatalf("expected newly uploaded image cleanup, got %q", deletedNew)
		}
	})
}

func TestProductHandlerDeleteProductImage(t *testing.T) {
	t.Run("returns 400 for invalid id", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.DeleteProductImage, http.MethodDelete, "/products/abc/image", nil, gin.Param{Key: "id", Value: "abc"})
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 404 when product is missing", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return nil, gorm.ErrRecordNotFound },
		}, &mockCategoryService{}, &mockFileService{})
		recorder := performJSONRequest(handler.DeleteProductImage, http.MethodDelete, "/products/1/image", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 and ignores image deletion error", func(t *testing.T) {
		handler := NewProductHandler(&mockProductService{
			getByIDFn: func(id uint) (*models.Product, error) { return &models.Product{Image: "old.png"}, nil },
			updateFn: func(id uint, updates map[string]interface{}) (*models.Product, error) {
				return &models.Product{Image: ""}, nil
			},
		}, &mockCategoryService{}, &mockFileService{
			deleteFn: func(filename string) error { return errors.New("fs error") },
		})
		recorder := performJSONRequest(handler.DeleteProductImage, http.MethodDelete, "/products/1/image", nil, gin.Param{Key: "id", Value: "1"})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}
