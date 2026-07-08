package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"poc-gin/services"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockAuthService struct {
	registerFn func(email, password string) (*models.User, error)
	loginFn    func(email, password string) (string, string, error)
	refreshFn  func(refreshToken string) (string, string, error)
	logoutFn   func(refreshToken string) error
}

func (m *mockAuthService) Register(_ context.Context, email, password string) (*models.User, error) {
	if m.registerFn != nil {
		return m.registerFn(email, password)
	}
	return nil, errors.New("unexpected Register call")
}

func (m *mockAuthService) Login(_ context.Context, email, password string) (string, string, error) {
	if m.loginFn != nil {
		return m.loginFn(email, password)
	}
	return "", "", errors.New("unexpected Login call")
}

func (m *mockAuthService) RefreshAccessToken(_ context.Context, refreshToken string) (string, string, error) {
	if m.refreshFn != nil {
		return m.refreshFn(refreshToken)
	}
	return "", "", errors.New("unexpected RefreshAccessToken call")
}

func (m *mockAuthService) Logout(_ context.Context, refreshToken string) error {
	if m.logoutFn != nil {
		return m.logoutFn(refreshToken)
	}
	return errors.New("unexpected Logout call")
}

type mockProductService struct {
	getAllFn           func(excludeSellerID *uint, page services.Pagination) ([]*models.Product, int64, error)
	getForSellerFn     func(sellerID uint) ([]*models.Product, error)
	getByIDFn          func(id uint) (*models.Product, error)
	getForManagementFn func(actorID uint, actorRole string, id uint) (*models.Product, error)
	createFn           func(product *models.Product) error
	updateFn           func(actorID uint, actorRole string, id uint, updates map[string]interface{}) (*models.Product, error)
	deleteFn           func(actorID uint, actorRole string, id uint) error
}

func (m *mockProductService) GetAllProducts(_ context.Context, excludeSellerID *uint, page services.Pagination) ([]*models.Product, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(excludeSellerID, page)
	}
	return nil, 0, errors.New("unexpected GetAllProducts call")
}

func (m *mockProductService) GetProductsForSeller(_ context.Context, sellerID uint) ([]*models.Product, error) {
	if m.getForSellerFn != nil {
		return m.getForSellerFn(sellerID)
	}
	return nil, errors.New("unexpected GetProductsForSeller call")
}

func (m *mockProductService) GetProductByID(_ context.Context, id uint) (*models.Product, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("unexpected GetProductByID call")
}

func (m *mockProductService) GetProductForManagement(_ context.Context, actorID uint, actorRole string, id uint) (*models.Product, error) {
	if m.getForManagementFn != nil {
		return m.getForManagementFn(actorID, actorRole, id)
	}
	return nil, errors.New("unexpected GetProductForManagement call")
}

func (m *mockProductService) CreateProduct(_ context.Context, product *models.Product) error {
	if m.createFn != nil {
		return m.createFn(product)
	}
	return errors.New("unexpected CreateProduct call")
}

func (m *mockProductService) UpdateProduct(_ context.Context, actorID uint, actorRole string, id uint, updates map[string]interface{}) (*models.Product, error) {
	if m.updateFn != nil {
		return m.updateFn(actorID, actorRole, id, updates)
	}
	return nil, errors.New("unexpected UpdateProduct call")
}

func (m *mockProductService) DeleteProduct(_ context.Context, actorID uint, actorRole string, id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(actorID, actorRole, id)
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

func (m *mockCategoryService) GetAllCategories(_ context.Context) ([]*models.Category, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, errors.New("unexpected GetAllCategories call")
}

func (m *mockCategoryService) GetCategoryByID(_ context.Context, id uint) (*models.Category, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("unexpected GetCategoryByID call")
}

func (m *mockCategoryService) CreateCategory(_ context.Context, category *models.Category) error {
	if m.createFn != nil {
		return m.createFn(category)
	}
	return errors.New("unexpected CreateCategory call")
}

func (m *mockCategoryService) UpdateCategory(_ context.Context, id uint, updates map[string]interface{}) (*models.Category, error) {
	if m.updateFn != nil {
		return m.updateFn(id, updates)
	}
	return nil, errors.New("unexpected UpdateCategory call")
}

func (m *mockCategoryService) DeleteCategory(_ context.Context, id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return errors.New("unexpected DeleteCategory call")
}

type mockPromotionService struct {
	getAllFn  func() ([]*models.Promotion, error)
	getByIDFn func(id uint) (*models.Promotion, error)
	createFn  func(input services.PromotionInput) (*models.Promotion, error)
	updateFn  func(id uint, input services.PromotionInput) (*models.Promotion, error)
	deleteFn  func(id uint) error
}

func (m *mockPromotionService) GetAllPromotions(_ context.Context) ([]*models.Promotion, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, errors.New("unexpected GetAllPromotions call")
}

func (m *mockPromotionService) GetPromotionByID(_ context.Context, id uint) (*models.Promotion, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("unexpected GetPromotionByID call")
}

func (m *mockPromotionService) CreatePromotion(_ context.Context, input services.PromotionInput) (*models.Promotion, error) {
	if m.createFn != nil {
		return m.createFn(input)
	}
	return nil, errors.New("unexpected CreatePromotion call")
}

func (m *mockPromotionService) UpdatePromotion(_ context.Context, id uint, input services.PromotionInput) (*models.Promotion, error) {
	if m.updateFn != nil {
		return m.updateFn(id, input)
	}
	return nil, errors.New("unexpected UpdatePromotion call")
}

func (m *mockPromotionService) DeletePromotion(_ context.Context, id uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return errors.New("unexpected DeletePromotion call")
}

type mockOrderService struct {
	createFn  func(userID uint, items []services.OrderItemInput) (*models.Order, error)
	listFn    func(userID uint, page services.Pagination) ([]*models.Order, int64, error)
	getByIDFn func(actorID, orderID uint, actorRole string) (*models.Order, error)
	updateFn  func(actorID, orderID uint, actorRole, status string) (*models.Order, error)
	deleteFn  func(actorID, orderID uint, actorRole string) error
}

func (m *mockOrderService) CreateOrder(_ context.Context, userID uint, items []services.OrderItemInput) (*models.Order, error) {
	if m.createFn != nil {
		return m.createFn(userID, items)
	}
	return nil, errors.New("unexpected CreateOrder call")
}

func (m *mockOrderService) GetOrdersForUser(_ context.Context, userID uint, page services.Pagination) ([]*models.Order, int64, error) {
	if m.listFn != nil {
		return m.listFn(userID, page)
	}
	return nil, 0, errors.New("unexpected GetOrdersForUser call")
}

func (m *mockOrderService) GetOrderByID(_ context.Context, actorID, orderID uint, actorRole string) (*models.Order, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(actorID, orderID, actorRole)
	}
	return nil, errors.New("unexpected GetOrderByID call")
}

func (m *mockOrderService) UpdateOrderStatus(_ context.Context, actorID, orderID uint, actorRole, status string) (*models.Order, error) {
	if m.updateFn != nil {
		return m.updateFn(actorID, orderID, actorRole, status)
	}
	return nil, errors.New("unexpected UpdateOrderStatus call")
}

func (m *mockOrderService) DeleteOrder(_ context.Context, actorID, orderID uint, actorRole string) error {
	if m.deleteFn != nil {
		return m.deleteFn(actorID, orderID, actorRole)
	}
	return errors.New("unexpected DeleteOrder call")
}

type mockOrderPaymentService struct {
	createCheckoutFn  func(actorID, orderID uint, actorRole, successURL, cancelURL string) (*services.OrderCheckoutSessionResult, error)
	releaseCheckoutFn func(actorID, orderID uint, actorRole string) error
	handleWebhookFn   func(payload []byte, signature string) error
}

func (m *mockOrderPaymentService) CreateStripeCheckoutSession(_ context.Context, actorID, orderID uint, actorRole, successURL, cancelURL string) (*services.OrderCheckoutSessionResult, error) {
	if m.createCheckoutFn != nil {
		return m.createCheckoutFn(actorID, orderID, actorRole, successURL, cancelURL)
	}
	return nil, errors.New("unexpected CreateStripeCheckoutSession call")
}

func (m *mockOrderPaymentService) ReleaseCheckoutSession(_ context.Context, actorID, orderID uint, actorRole string) error {
	if m.releaseCheckoutFn != nil {
		return m.releaseCheckoutFn(actorID, orderID, actorRole)
	}
	// Deleting an order always releases its checkout first; default to a
	// no-op so existing deletion tests keep exercising the delete path.
	return nil
}

func (m *mockOrderPaymentService) HandleStripeWebhook(_ context.Context, payload []byte, signature string) error {
	if m.handleWebhookFn != nil {
		return m.handleWebhookFn(payload, signature)
	}
	return errors.New("unexpected HandleStripeWebhook call")
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

func performAuthenticatedJSONRequest(handlerFunc gin.HandlerFunc, method, target string, body any, params ...gin.Param) *httptest.ResponseRecorder {
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
	ctx.Set(constants.ContextKeyUserID, uint(1))
	ctx.Set(constants.ContextKeyUserRole, constants.RoleUser)
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
