package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubPinger struct {
	err error
}

func (s *stubPinger) Ping(_ context.Context) error {
	return s.err
}

func performHealthRequest(handlerFunc gin.HandlerFunc, target string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req, _ := http.NewRequest(http.MethodGet, target, nil)
	ctx.Request = req
	handlerFunc(ctx)
	return recorder
}

func TestHealthHandlerHealthzAlwaysOK(t *testing.T) {
	handler := NewHealthHandler(&stubPinger{err: errors.New("db is down")})

	recorder := performHealthRequest(handler.Healthz, "/healthz")
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 regardless of db state, got %d", recorder.Code)
	}
}

func TestHealthHandlerReadyz(t *testing.T) {
	t.Run("returns 200 when database is reachable", func(t *testing.T) {
		handler := NewHealthHandler(&stubPinger{})

		recorder := performHealthRequest(handler.Readyz, "/readyz")
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})

	t.Run("returns 503 when database is unreachable", func(t *testing.T) {
		handler := NewHealthHandler(&stubPinger{err: errors.New("connection refused")})

		recorder := performHealthRequest(handler.Readyz, "/readyz")
		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", recorder.Code)
		}
	})
}
