package database

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"poc-gin/config"
)

func testDatabaseConfig() *config.DatabaseConfig {
	env := func(key, fallback string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return fallback
	}

	return &config.DatabaseConfig{
		Host:     env("DB_HOST", "127.0.0.1"),
		Port:     env("DB_PORT", "5432"),
		User:     env("DB_USER", "golang"),
		Password: env("DB_PASSWORD", "golang"),
		Name:     env("DB_NAME", "ecommerce"),
	}
}

func TestNewConnectsPingsAndCloses(t *testing.T) {
	cfg := testDatabaseConfig()

	db, err := New(cfg)
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("postgres is required in CI: %v", err)
		}
		t.Skipf("postgres not available: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("expected ping success, got %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("expected close success, got %v", err)
	}
}

func TestNewReturnsFormattedErrorOnConnectionFailure(t *testing.T) {
	cfg := testDatabaseConfig()
	cfg.Port = "1"

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected connection failure on unreachable port")
	}
	if !strings.Contains(err.Error(), "failed to ping database") &&
		!strings.Contains(err.Error(), "failed to connect to database") {
		t.Fatalf("expected formatted connection error, got %v", err)
	}
}

func TestConnectionHint(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host: "127.0.0.1",
		Port: "5432",
	}

	t.Run("adds reset guidance for password auth failures", func(t *testing.T) {
		err := errors.New(`failed SASL auth: FATAL: password authentication failed for user "golang" (SQLSTATE 28P01)`)

		hint := connectionHint(cfg, err)

		if !strings.Contains(hint, "docker compose down -v") {
			t.Fatalf("expected volume reset guidance, got %q", hint)
		}
		if !strings.Contains(hint, "DB_USER") || !strings.Contains(hint, "DB_PASSWORD") {
			t.Fatalf("expected credential guidance, got %q", hint)
		}
	})

	t.Run("adds startup guidance for connection refused errors", func(t *testing.T) {
		err := errors.New(`dial tcp 127.0.0.1:5432: connectex: No connection could be made because the target machine actively refused it`)

		hint := connectionHint(cfg, err)

		if !strings.Contains(hint, "docker compose up -d db") {
			t.Fatalf("expected docker compose startup guidance, got %q", hint)
		}
		if !strings.Contains(hint, "127.0.0.1:5432") {
			t.Fatalf("expected host and port in hint, got %q", hint)
		}
	})
}
