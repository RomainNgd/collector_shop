package config

import (
	"os"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Run("fails when database configuration is incomplete", func(t *testing.T) {
		cfg := &Config{}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("fails when jwt secret is missing", func(t *testing.T) {
		cfg := &Config{
			Database: DatabaseConfig{
				Host:     "127.0.0.1",
				Port:     "5432",
				User:     "user",
				Password: "password",
				Name:     "db",
			},
		}
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected jwt validation error")
		}
	})

	t.Run("passes with complete config", func(t *testing.T) {
		cfg := &Config{
			Database: DatabaseConfig{
				Host:     "127.0.0.1",
				Port:     "5432",
				User:     "user",
				Password: "password",
				Name:     "db",
			},
			JWT: JWTConfig{Secret: "secret"},
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected valid config, got %v", err)
		}
	})
}

func TestGetEnvAsInt64(t *testing.T) {
	key := "TEST_MAX_FILE_SIZE"
	t.Cleanup(func() { _ = os.Unsetenv(key) })

	if err := os.Setenv(key, "1234"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if got := getEnvAsInt64(key, 1); got != 1234 {
		t.Fatalf("expected 1234, got %d", got)
	}

	if err := os.Setenv(key, "invalid"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if got := getEnvAsInt64(key, 99); got != 99 {
		t.Fatalf("expected fallback 99, got %d", got)
	}
}

func TestLoad(t *testing.T) {
	keys := []string{
		"PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"DB_AUTO_MIGRATE", "JWT_SECRET", "UPLOAD_DIR", "MAX_FILE_SIZE",
	}

	for _, key := range keys {
		original, hadValue := os.LookupEnv(key)
		t.Cleanup(func() {
			if hadValue {
				_ = os.Setenv(key, original)
			} else {
				_ = os.Unsetenv(key)
			}
		})
		_ = os.Unsetenv(key)
	}

	env := map[string]string{
		"DB_HOST":         "127.0.0.1",
		"DB_PORT":         "5432",
		"DB_USER":         "user",
		"DB_PASSWORD":     "password",
		"DB_NAME":         "db",
		"JWT_SECRET":      "secret",
		"DB_AUTO_MIGRATE": "true",
	}
	for key, value := range env {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set env %s: %v", key, err)
		}
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Server.Port)
	}
	if cfg.Upload.Dir != "./upload" {
		t.Fatalf("expected default upload dir, got %s", cfg.Upload.Dir)
	}
	if cfg.Upload.MaxFileSize != 5242880 {
		t.Fatalf("expected default max file size, got %d", cfg.Upload.MaxFileSize)
	}
	if !cfg.Database.AutoMigrate {
		t.Fatal("expected auto migrate to be true")
	}
}
