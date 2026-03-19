package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"poc-gin/models"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockAuthService struct {
	registerFn func(email, password string) (*models.User, error)
	loginFn    func(email, password string) (string, error)
}

func (m *mockAuthService) Register(email, password string) (*models.User, error) {
	if m.registerFn != nil {
		return m.registerFn(email, password)
	}
	return nil, errors.New("unexpected Register call")
}

func (m *mockAuthService) Login(email, password string) (string, error) {
	if m.loginFn != nil {
		return m.loginFn(email, password)
	}
	return "", errors.New("unexpected Login call")
}

type mockProductService struct {
	getAllFn  func() ([]*models.Product, error)
	getByIDFn func(id uint) (*models.Product, error)
	createFn  func(product *models.Product) error
	updateFn  func(id uint, updates map[string]interface{}) (*models.Product, error)
	deleteFn  func(id uint) error
}

func (m *mockProductService) GetAllProducts() ([]*models.Product, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, errors.New("unexpected GetAllProducts call")
}

func (m *mockProductService) GetProductByID(id uint) (*models.Product, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("unexpected GetProductByID call")
}

func (m *mockProductService) CreateProduct(product *models.Product) error {
	if m.createFn != nil {
		return m.createFn(product)
	}
	return errors.New("unexpected CreateProduct call")
}

func (m *mockProductService) UpdateProduct(id uint, updates map[string]interface{}) (*models.Product, error) {
	if m.updateFn != nil {
		return m.updateFn(id, updates)
	}
	return nil, errors.New("unexpected UpdateProduct call")
}

func (m *mockProductService) DeleteProduct(id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return errors.New("unexpected DeleteProduct call")
}

type mockCategoryService struct {
	getAllFn  func() ([]*models.Category, error)
	getByIDFn func(id uint) (*models.Category, error)
	createFn  func(category *models.Category) error
	updateFn  func(id uint, updates map[string]interface{}) (*models.Category, error)
	deleteFn  func(id uint) error
}

func (m *mockCategoryService) GetAllCategories() ([]*models.Category, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, errors.New("unexpected GetAllCategories call")
}

func (m *mockCategoryService) GetCategoryByID(id uint) (*models.Category, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("unexpected GetCategoryByID call")
}

func (m *mockCategoryService) CreateCategory(category *models.Category) error {
	if m.createFn != nil {
		return m.createFn(category)
	}
	return errors.New("unexpected CreateCategory call")
}

func (m *mockCategoryService) UpdateCategory(id uint, updates map[string]interface{}) (*models.Category, error) {
	if m.updateFn != nil {
		return m.updateFn(id, updates)
	}
	return nil, errors.New("unexpected UpdateCategory call")
}

func (m *mockCategoryService) DeleteCategory(id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return errors.New("unexpected DeleteCategory call")
}

type mockFileService struct {
	saveFn   func(file *multipart.FileHeader) (string, error)
	deleteFn func(filename string) error
}

func (m *mockFileService) SaveImage(file *multipart.FileHeader) (string, error) {
	if m.saveFn != nil {
		return m.saveFn(file)
	}
	return "", errors.New("unexpected SaveImage call")
}

func (m *mockFileService) DeleteImage(filename string) error {
	if m.deleteFn != nil {
		return m.deleteFn(filename)
	}
	return errors.New("unexpected DeleteImage call")
}

func performJSONRequest(handlerFunc gin.HandlerFunc, method, target string, body any, params ...gin.Param) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}

	req, _ := http.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Params = params

	handlerFunc(ctx)
	return recorder
}

func decodeAPIResponse(recorder *httptest.ResponseRecorder, t interface {
	Helper()
	Fatalf(string, ...interface{})
}) APIResponse {
	t.Helper()

	var resp APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}
