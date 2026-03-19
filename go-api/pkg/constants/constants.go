package constants

import "time"

const (
	// Roles
	RoleUser  = "ROLE_USER"
	RoleAdmin = "ROLE_ADMIN"

	// Context keys
	ContextKeyUserID   = "userID"
	ContextKeyUserRole = "userRole"

	// JWT
	JWTExpirationHours = 24

	// Database
	DBTimeout = 5 * time.Second

	// File upload
	MaxFileSize        = 5 << 20 // 5MB
	UploadDir          = "./upload"
	AllowedImageExts   = ".jpg,.jpeg,.png,.webp"
)
