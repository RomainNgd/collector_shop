package database

import (
	"errors"
	"strings"
	"testing"

	"poc-gin/config"
)

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
