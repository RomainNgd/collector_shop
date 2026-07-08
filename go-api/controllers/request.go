package controllers

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CreateProductRequest struct {
	Name            string  `json:"name" binding:"required,min=2,max=120"`
	Description     string  `json:"description" binding:"required,min=2,max=1000"`
	Image           string  `json:"image" binding:"omitempty,max=255"`
	Price           float64 `json:"price" binding:"required,gt=0"`
	Stock           int     `json:"stock" binding:"required,gt=0"`
	CategoryID      uint    `json:"category_id" binding:"required,gt=0"`
	PromotionType   string  `json:"promotion_type" binding:"omitempty,oneof=percentage fixed"`
	PromotionValue  float64 `json:"promotion_value" binding:"omitempty,gte=0"`
	PromotionActive *bool   `json:"promotion_active"`
}

type UpdateProductRequest struct {
	Name            string  `json:"name" binding:"required,min=2,max=120"`
	Description     string  `json:"description" binding:"required,min=2,max=1000"`
	Image           string  `json:"image" binding:"omitempty,max=255"`
	Price           float64 `json:"price" binding:"required,gt=0"`
	Stock           int     `json:"stock" binding:"required,gt=0"`
	IsActive        *bool   `json:"is_active" binding:"required"`
	CategoryID      uint    `json:"category_id" binding:"required,gt=0"`
	PromotionType   string  `json:"promotion_type" binding:"omitempty,oneof=percentage fixed"`
	PromotionValue  float64 `json:"promotion_value" binding:"omitempty,gte=0"`
	PromotionActive *bool   `json:"promotion_active" binding:"required"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=120"`
	Description string `json:"description" binding:"required,min=2,max=1000"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=120"`
	Description string `json:"description" binding:"required,min=2,max=1000"`
}

type CreatePromotionRequest struct {
	Name         string  `json:"name" binding:"required,min=2,max=120"`
	Description  string  `json:"description" binding:"omitempty,max=1000"`
	Type         string  `json:"type" binding:"required,oneof=percentage fixed"`
	Value        float64 `json:"value" binding:"required,gt=0"`
	IsActive     *bool   `json:"is_active" binding:"required"`
	AppliesToAll *bool   `json:"applies_to_all" binding:"required"`
	ProductIDs   []uint  `json:"product_ids"`
}

type UpdatePromotionRequest struct {
	Name         string  `json:"name" binding:"required,min=2,max=120"`
	Description  string  `json:"description" binding:"omitempty,max=1000"`
	Type         string  `json:"type" binding:"required,oneof=percentage fixed"`
	Value        float64 `json:"value" binding:"required,gt=0"`
	IsActive     *bool   `json:"is_active" binding:"required"`
	AppliesToAll *bool   `json:"applies_to_all" binding:"required"`
	ProductIDs   []uint  `json:"product_ids"`
}

type CreateOrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required,gt=0"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	Items []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateOrderRequest struct {
	Status string `json:"status" binding:"required"`
}

type CreateOrderCheckoutSessionRequest struct {
	SuccessURL string `json:"success_url" binding:"required"`
	CancelURL  string `json:"cancel_url" binding:"required"`
}
