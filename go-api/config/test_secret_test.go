package config

import (
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func newTestSecret(t *testing.T) string {
	t.Helper()

	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		t.Fatalf("failed to generate test secret: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(value)
}
