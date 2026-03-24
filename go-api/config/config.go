package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
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

	return nil
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
