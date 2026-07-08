package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"testing"
	"time"

	"gorm.io/gorm"
)

func seedProduct(t *testing.T, tx *gorm.DB, categoryID uint) *models.Product {
	t.Helper()

	seller := seedUser(t, tx, "ROLE_USER")
	product := &models.Product{
		Name:        fmt.Sprintf("Product-%d", time.Now().UnixNano()),
		Description: "Test product",
		Image:       "image.png",
		Price:       10.5,
		Stock:       10,
		IsActive:    true,
		SellerID:    &seller.ID,
		CategoryID:  categoryID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return product
}

func TestProductServiceCRUDAndPreload(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear global promotions: %v", err)
	}
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, "ROLE_USER")
	service := NewProductService(tx)

	product := &models.Product{
		Name:        fmt.Sprintf("Blue-Eyes-%d", time.Now().UnixNano()),
		Description: "Mint card",
		Image:       "blue-eyes.png",
		Price:       19.99,
		Stock:       5,
		IsActive:    true,
		SellerID:    &seller.ID,
		CategoryID:  category.ID,
	}

	if err := service.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if product.Category.ID != category.ID {
		t.Fatalf("expected preloaded category %d, got %#v", category.ID, product.Category)
	}
	if product.EffectivePrice != product.Price || product.AppliedPromotion != nil {
		t.Fatalf("expected no promotion on create, got effective=%f promotion=%#v", product.EffectivePrice, product.AppliedPromotion)
	}

	found, err := service.GetProductByID(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("expected get success, got %v", err)
	}
	if found.Category.ID != category.ID {
		t.Fatalf("expected preloaded category %d, got %#v", category.ID, found.Category)
	}
	if found.EffectivePrice != found.Price || found.AppliedPromotion != nil {
		t.Fatalf("expected no promotion on get, got effective=%f promotion=%#v", found.EffectivePrice, found.AppliedPromotion)
	}

	all, total, err := service.GetAllProducts(context.Background(), nil, Pagination{})
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one product")
	}
	if total != int64(len(all)) {
		t.Fatalf("expected total %d to match returned count %d", total, len(all))
	}

	updated, err := service.UpdateProduct(context.Background(), seller.ID, seller.Role, product.ID, map[string]interface{}{
		"name":  product.Name + "-updated",
		"price": 29.99,
		"stock": 7,
	})
	if err != nil {
		t.Fatalf("expected update success, got %v", err)
	}
	if updated.Price != 29.99 {
		t.Fatalf("expected updated price, got %f", updated.Price)
	}
	if updated.Category.ID != category.ID {
		t.Fatalf("expected preloaded category after update, got %#v", updated.Category)
	}
	if updated.EffectivePrice != updated.Price || updated.AppliedPromotion != nil {
		t.Fatalf("expected no promotion on update, got effective=%f promotion=%#v", updated.EffectivePrice, updated.AppliedPromotion)
	}

	if err := service.DeleteProduct(context.Background(), seller.ID, seller.Role, product.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
}

func TestProductServiceAppliesSellerPromotion(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, "ROLE_USER")
	service := NewProductService(tx)

	productA := &models.Product{
		Name:            fmt.Sprintf("Console-%d", time.Now().UnixNano()),
		Description:     "Limited edition",
		Image:           "console.png",
		Price:           100,
		Stock:           3,
		IsActive:        true,
		SellerID:        &seller.ID,
		CategoryID:      category.ID,
		PromotionType:   models.PromotionTypePercentage,
		PromotionValue:  20,
		PromotionActive: true,
	}
	if err := tx.Create(productA).Error; err != nil {
		t.Fatalf("failed to create product A: %v", err)
	}

	foundA, err := service.GetProductByID(context.Background(), productA.ID)
	if err != nil {
		t.Fatalf("expected product A lookup success, got %v", err)
	}
	if foundA.EffectivePrice != 80 {
		t.Fatalf("expected product A effective price 80, got %f", foundA.EffectivePrice)
	}
	if foundA.AppliedPromotion == nil || foundA.AppliedPromotion.Type != models.PromotionTypePercentage {
		t.Fatalf("expected seller promotion for product A, got %#v", foundA.AppliedPromotion)
	}
}

func TestProductServiceGetAllProductsExcludesOwnSeller(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	service := NewProductService(tx)

	ownProduct := seedProduct(t, tx, category.ID)
	otherProduct := seedProduct(t, tx, category.ID)

	all, _, err := service.GetAllProducts(context.Background(), ownProduct.SellerID, Pagination{})
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}

	for _, product := range all {
		if product.ID == ownProduct.ID {
			t.Fatalf("expected own product %d to be excluded from catalog", ownProduct.ID)
		}
	}

	found := false
	for _, product := range all {
		if product.ID == otherProduct.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected other seller's product %d to remain in catalog", otherProduct.ID)
	}
}

func TestProductServiceGetAllProductsPagination(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	service := NewProductService(tx)

	for i := 0; i < 3; i++ {
		seedProduct(t, tx, category.ID)
	}

	firstPage, total, err := service.GetAllProducts(context.Background(), nil, Pagination{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("expected first page success, got %v", err)
	}
	if total < 3 {
		t.Fatalf("expected total of at least 3, got %d", total)
	}
	if len(firstPage) != 2 {
		t.Fatalf("expected 2 products on first page, got %d", len(firstPage))
	}

	secondPage, totalAgain, err := service.GetAllProducts(context.Background(), nil, Pagination{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("expected second page success, got %v", err)
	}
	if totalAgain != total {
		t.Fatalf("expected stable total across pages, got %d then %d", total, totalAgain)
	}
	if len(secondPage) == 0 {
		t.Fatal("expected at least one product on second page")
	}

	firstIDs := make(map[uint]bool, len(firstPage))
	for _, p := range firstPage {
		firstIDs[p.ID] = true
	}
	for _, p := range secondPage {
		if firstIDs[p.ID] {
			t.Fatalf("expected no overlap between pages, product %d appeared twice", p.ID)
		}
	}
}

func TestProductServiceGetAllProductsLimitIsCapped(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	service := NewProductService(tx)

	for i := 0; i < MaxPageLimit+5; i++ {
		seedProduct(t, tx, category.ID)
	}

	page, total, err := service.GetAllProducts(context.Background(), nil, Pagination{Limit: 100000})
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if total < int64(MaxPageLimit+5) {
		t.Fatalf("expected total to reflect all seeded products, got %d", total)
	}
	if len(page) != MaxPageLimit {
		t.Fatalf("expected page size capped at %d, got %d", MaxPageLimit, len(page))
	}
}

func TestProductServiceUpdateNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	user := seedUser(t, tx, "ROLE_USER")

	_, err := service.UpdateProduct(context.Background(), user.ID, user.Role, 999999, map[string]interface{}{"name": "x"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestProductServiceDeleteNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	user := seedUser(t, tx, "ROLE_USER")

	err := service.DeleteProduct(context.Background(), user.ID, user.Role, 999999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestProductServiceGetProductsForSeller(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, constants.RoleUser)
	otherSeller := seedUser(t, tx, constants.RoleUser)
	service := NewProductService(tx)

	product := &models.Product{
		Name:        fmt.Sprintf("Seller-Product-%d", time.Now().UnixNano()),
		Description: "For sale",
		Image:       "image.png",
		Price:       15,
		Stock:       2,
		IsActive:    true,
		SellerID:    &seller.ID,
		CategoryID:  category.ID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed seller product: %v", err)
	}
	otherProduct := &models.Product{
		Name:        fmt.Sprintf("Other-Seller-Product-%d", time.Now().UnixNano()),
		Description: "Not for this seller",
		Image:       "image.png",
		Price:       15,
		Stock:       2,
		IsActive:    true,
		SellerID:    &otherSeller.ID,
		CategoryID:  category.ID,
	}
	if err := tx.Create(otherProduct).Error; err != nil {
		t.Fatalf("failed to seed other seller product: %v", err)
	}

	products, err := service.GetProductsForSeller(context.Background(), seller.ID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected exactly one product for seller, got %d", len(products))
	}
	if products[0].ID != product.ID {
		t.Fatalf("expected product %d, got %d", product.ID, products[0].ID)
	}
	if products[0].SellerEmail != seller.Email {
		t.Fatalf("expected seller email populated, got %q", products[0].SellerEmail)
	}
}

func TestProductServiceCreateProductRejectsInvalidData(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, constants.RoleUser)
	service := NewProductService(tx)

	baseProduct := func() *models.Product {
		return &models.Product{
			Name:        fmt.Sprintf("Invalid-%d", time.Now().UnixNano()),
			Description: "Test",
			Image:       "image.png",
			Price:       10,
			Stock:       5,
			IsActive:    true,
			SellerID:    &seller.ID,
			CategoryID:  category.ID,
		}
	}

	t.Run("missing seller", func(t *testing.T) {
		product := baseProduct()
		product.SellerID = nil
		err := service.CreateProduct(context.Background(), product)
		if !errors.Is(err, ErrProductSellerRequired) {
			t.Fatalf("expected ErrProductSellerRequired, got %v", err)
		}
	})

	t.Run("invalid stock", func(t *testing.T) {
		product := baseProduct()
		product.Stock = 0
		err := service.CreateProduct(context.Background(), product)
		if !errors.Is(err, ErrProductInvalidStock) {
			t.Fatalf("expected ErrProductInvalidStock, got %v", err)
		}
	})

	t.Run("invalid promotion type", func(t *testing.T) {
		product := baseProduct()
		product.PromotionActive = true
		product.PromotionType = "unknown"
		product.PromotionValue = 10
		err := service.CreateProduct(context.Background(), product)
		if !errors.Is(err, ErrProductInvalidPromotionType) {
			t.Fatalf("expected ErrProductInvalidPromotionType, got %v", err)
		}
	})

	t.Run("invalid promotion value", func(t *testing.T) {
		product := baseProduct()
		product.PromotionActive = true
		product.PromotionType = models.PromotionTypePercentage
		product.PromotionValue = 150
		err := service.CreateProduct(context.Background(), product)
		if !errors.Is(err, ErrProductInvalidPromotionValue) {
			t.Fatalf("expected ErrProductInvalidPromotionValue, got %v", err)
		}
	})
}

func TestProductServiceUpdateProductAccessControl(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, constants.RoleUser)
	otherUser := seedUser(t, tx, constants.RoleUser)
	admin := seedUser(t, tx, constants.RoleAdmin)
	product := seedProduct(t, tx, category.ID)
	product.SellerID = &seller.ID
	if err := tx.Save(product).Error; err != nil {
		t.Fatalf("failed to assign seller: %v", err)
	}
	service := NewProductService(tx)

	t.Run("owner can update", func(t *testing.T) {
		updated, err := service.UpdateProduct(context.Background(), seller.ID, seller.Role, product.ID, map[string]interface{}{"stock": 4})
		if err != nil {
			t.Fatalf("expected owner update success, got %v", err)
		}
		if updated.Stock != 4 {
			t.Fatalf("expected stock 4, got %d", updated.Stock)
		}
	})

	t.Run("admin can update", func(t *testing.T) {
		if _, err := service.UpdateProduct(context.Background(), admin.ID, admin.Role, product.ID, map[string]interface{}{"stock": 6}); err != nil {
			t.Fatalf("expected admin update success, got %v", err)
		}
	})

	t.Run("other user is denied", func(t *testing.T) {
		_, err := service.UpdateProduct(context.Background(), otherUser.ID, otherUser.Role, product.ID, map[string]interface{}{"stock": 8})
		if !errors.Is(err, ErrProductAccessDenied) {
			t.Fatalf("expected ErrProductAccessDenied, got %v", err)
		}
	})

	t.Run("invalid stock update rejected", func(t *testing.T) {
		_, err := service.UpdateProduct(context.Background(), seller.ID, seller.Role, product.ID, map[string]interface{}{"stock": "not-a-number"})
		if !errors.Is(err, ErrProductInvalidStock) {
			t.Fatalf("expected ErrProductInvalidStock, got %v", err)
		}
	})

	t.Run("invalid promotion active update rejected", func(t *testing.T) {
		_, err := service.UpdateProduct(context.Background(), seller.ID, seller.Role, product.ID, map[string]interface{}{"promotion_active": "not-a-bool"})
		if !errors.Is(err, ErrProductInvalidPromotionValue) {
			t.Fatalf("expected ErrProductInvalidPromotionValue, got %v", err)
		}
	})

	t.Run("invalid promotion value update rejected", func(t *testing.T) {
		_, err := service.UpdateProduct(context.Background(), seller.ID, seller.Role, product.ID, map[string]interface{}{
			"promotion_active": true,
			"promotion_type":   models.PromotionTypeFixed,
			"promotion_value":  "not-a-number",
		})
		if !errors.Is(err, ErrProductInvalidPromotionValue) {
			t.Fatalf("expected ErrProductInvalidPromotionValue, got %v", err)
		}
	})
}

func TestProductServiceDeleteProductAccessControl(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, constants.RoleUser)
	otherUser := seedUser(t, tx, constants.RoleUser)
	product := seedProduct(t, tx, category.ID)
	product.SellerID = &seller.ID
	if err := tx.Save(product).Error; err != nil {
		t.Fatalf("failed to assign seller: %v", err)
	}
	service := NewProductService(tx)

	err := service.DeleteProduct(context.Background(), otherUser.ID, otherUser.Role, product.ID)
	if !errors.Is(err, ErrProductAccessDenied) {
		t.Fatalf("expected ErrProductAccessDenied, got %v", err)
	}

	if err := service.DeleteProduct(context.Background(), seller.ID, seller.Role, product.ID); err != nil {
		t.Fatalf("expected owner delete success, got %v", err)
	}
}

func TestCanManageProduct(t *testing.T) {
	sellerID := uint(5)
	tests := []struct {
		name     string
		actorID  uint
		role     string
		sellerID *uint
		want     bool
	}{
		{"admin can always manage", 1, constants.RoleAdmin, nil, true},
		{"owner can manage", 5, constants.RoleUser, &sellerID, true},
		{"non-owner cannot manage", 6, constants.RoleUser, &sellerID, false},
		{"nil seller cannot be managed by user", 5, constants.RoleUser, nil, false},
		{"zero actor id cannot manage", 0, constants.RoleUser, &sellerID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canManageProduct(tt.actorID, tt.role, tt.sellerID); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestPopulateProductSellerEmail(t *testing.T) {
	populateProductSellerEmail(nil)

	product := &models.Product{Seller: models.User{Email: "seller@example.com"}}
	populateProductSellerEmail(product)
	if product.SellerEmail != "seller@example.com" {
		t.Fatalf("expected seller email populated, got %q", product.SellerEmail)
	}
}

func TestValidateProductPromotion(t *testing.T) {
	tests := []struct {
		name          string
		active        bool
		promotionType string
		value         float64
		wantErr       error
	}{
		{"inactive skips validation", false, "unknown", -5, nil},
		{"valid percentage", true, models.PromotionTypePercentage, 50, nil},
		{"percentage too high", true, models.PromotionTypePercentage, 101, ErrProductInvalidPromotionValue},
		{"percentage zero invalid", true, models.PromotionTypePercentage, 0, ErrProductInvalidPromotionValue},
		{"valid fixed", true, models.PromotionTypeFixed, 10, nil},
		{"fixed zero invalid", true, models.PromotionTypeFixed, 0, ErrProductInvalidPromotionValue},
		{"unknown type invalid", true, "unknown", 10, ErrProductInvalidPromotionType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProductPromotion(tt.active, tt.promotionType, tt.value)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestIntFromUpdate(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		want   int
		wantOK bool
	}{
		{"int", 5, 5, true},
		{"uint", uint(7), 7, true},
		{"uint overflow", ^uint(0), 0, false},
		{"unsupported type", "5", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intFromUpdate(tt.value)
			if ok != tt.wantOK || (ok && got != tt.want) {
				t.Fatalf("expected (%d, %v), got (%d, %v)", tt.want, tt.wantOK, got, ok)
			}
		})
	}
}

func TestFloatFromUpdate(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		want   float64
		wantOK bool
	}{
		{"float64", float64(1.5), 1.5, true},
		{"float32", float32(2.5), 2.5, true},
		{"int", 3, 3, true},
		{"unsupported type", "3", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := floatFromUpdate(tt.value)
			if ok != tt.wantOK || (ok && got != tt.want) {
				t.Fatalf("expected (%f, %v), got (%f, %v)", tt.want, tt.wantOK, got, ok)
			}
		})
	}
}

func TestBoolFromUpdate(t *testing.T) {
	if v, ok := boolFromUpdate(true); !ok || !v {
		t.Fatalf("expected (true, true), got (%v, %v)", v, ok)
	}
	if v, ok := boolFromUpdate("true"); ok || v {
		t.Fatalf("expected (false, false), got (%v, %v)", v, ok)
	}
}

func TestCategoryDeletionRestrictedWhenProductReferencesIt(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	_ = seedProduct(t, tx, category.ID)
	categoryService := NewCategoryService(tx)

	err := categoryService.DeleteCategory(context.Background(), category.ID)
	if !errors.Is(err, ErrCategoryInUse) {
		t.Fatalf("expected ErrCategoryInUse, got %v", err)
	}
}

func TestProductServiceGetProductForManagement(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	service := NewProductService(tx)

	t.Run("owner can manage their own product", func(t *testing.T) {
		product := seedProduct(t, tx, category.ID)

		found, err := service.GetProductForManagement(context.Background(), *product.SellerID, constants.RoleUser, product.ID)
		if err != nil {
			t.Fatalf("expected success for owner, got %v", err)
		}
		if found.ID != product.ID {
			t.Fatalf("expected product %d, got %d", product.ID, found.ID)
		}
		if found.Category.ID != category.ID {
			t.Fatalf("expected preloaded category, got %#v", found.Category)
		}
	})

	t.Run("admin can manage any product", func(t *testing.T) {
		product := seedProduct(t, tx, category.ID)
		admin := seedUser(t, tx, constants.RoleAdmin)

		found, err := service.GetProductForManagement(context.Background(), admin.ID, constants.RoleAdmin, product.ID)
		if err != nil {
			t.Fatalf("expected success for admin, got %v", err)
		}
		if found.ID != product.ID {
			t.Fatalf("expected product %d, got %d", product.ID, found.ID)
		}
	})

	t.Run("non-owner is denied", func(t *testing.T) {
		product := seedProduct(t, tx, category.ID)
		other := seedUser(t, tx, constants.RoleUser)

		_, err := service.GetProductForManagement(context.Background(), other.ID, constants.RoleUser, product.ID)
		if !errors.Is(err, ErrProductAccessDenied) {
			t.Fatalf("expected ErrProductAccessDenied, got %v", err)
		}
	})

	t.Run("owner can manage an inactive out-of-stock product", func(t *testing.T) {
		product := seedProduct(t, tx, category.ID)
		if err := tx.Model(product).Updates(map[string]interface{}{"is_active": false, "stock": 0}).Error; err != nil {
			t.Fatalf("failed to deactivate product: %v", err)
		}

		found, err := service.GetProductForManagement(context.Background(), *product.SellerID, constants.RoleUser, product.ID)
		if err != nil {
			t.Fatalf("expected management access to bypass catalog filters, got %v", err)
		}
		if found.IsActive {
			t.Fatal("expected fetched product to reflect inactive state")
		}
	})

	t.Run("returns not found for unknown product", func(t *testing.T) {
		user := seedUser(t, tx, constants.RoleUser)

		_, err := service.GetProductForManagement(context.Background(), user.ID, constants.RoleUser, 9999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected ErrRecordNotFound, got %v", err)
		}
	})
}
