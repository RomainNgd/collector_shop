package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
	Stripe   StripeConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	Name        string
	AutoMigrate bool
}

type JWTConfig struct {
	Secret string
}

type UploadConfig struct {
	Dir         string
	MaxFileSize int64
}

type StripeConfig struct {
	Enabled                bool
	SecretKey              string
	WebhookSecret          string
	CheckoutAllowedOrigins []string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  .env file not found, using system environment variables")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "127.0.0.1"),
			Port:        getEnv("DB_PORT", "5432"),
			User:        getEnv("DB_USER", "golang"),
			Password:    getEnv("DB_PASSWORD", "golang"),
			Name:        getEnv("DB_NAME", "ecommerce"),
			AutoMigrate: getEnv("DB_AUTO_MIGRATE", "false") == "true",
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		Upload: UploadConfig{
			Dir:         getEnv("UPLOAD_DIR", "./upload"),
			MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", 5242880), // 5MB default
		},
		Stripe: StripeConfig{
			Enabled:                getEnv("STRIPE_ENABLED", "false") == "true",
			SecretKey:              getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecret:          getEnv("STRIPE_WEBHOOK_SECRET", ""),
			CheckoutAllowedOrigins: getEnvList("STRIPE_CHECKOUT_ALLOWED_ORIGINS"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Database.Host == "" || c.Database.Port == "" ||
		c.Database.User == "" || c.Database.Password == "" ||
		c.Database.Name == "" {
		return fmt.Errorf("missing required database configuration")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.Stripe.Enabled {
		if c.Stripe.SecretKey == "" {
			return fmt.Errorf("STRIPE_SECRET_KEY is required when STRIPE_ENABLED=true")
		}

		if c.Stripe.WebhookSecret == "" {
			return fmt.Errorf("STRIPE_WEBHOOK_SECRET is required when STRIPE_ENABLED=true")
		}

		if len(c.Stripe.CheckoutAllowedOrigins) == 0 {
			return fmt.Errorf("STRIPE_CHECKOUT_ALLOWED_ORIGINS is required when STRIPE_ENABLED=true")
		}
	}

	return nil
}

func getEnvList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}
