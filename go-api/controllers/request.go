package controllers

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=120"`
	Description string  `json:"description" binding:"required,min=2,max=1000"`
	Image       string  `json:"image" binding:"required,min=2,max=120"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	CategoryID  uint    `json:"category_id" binding:"required,gt=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=120"`
	Description string  `json:"description" binding:"required,min=2,max=1000"`
	Image       string  `json:"image" binding:"required,min=2,max=120"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	CategoryID  uint    `json:"category_id" binding:"required,gt=0"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=120"`
	Description string `json:"description" binding:"required,min=2,max=1000"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=120"`
	Description string `json:"description" binding:"required,min=2,max=1000"`
}
