package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newRateLimitedRouter(limiter *RateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/login", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func performRateLimitedRequest(router *gin.Engine, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = ip + ":12345"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)
	router := newRateLimitedRouter(limiter)

	for i := 0; i < 3; i++ {
		if recorder := performRateLimitedRequest(router, "10.0.0.1"); recorder.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, recorder.Code)
		}
	}

	recorder := performRateLimitedRequest(router, "10.0.0.1")
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after limit, got %d", recorder.Code)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header on 429 response")
	}
}

func TestRateLimiterIsolatesClients(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	router := newRateLimitedRouter(limiter)

	if recorder := performRateLimitedRequest(router, "10.0.0.1"); recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for first client, got %d", recorder.Code)
	}
	if recorder := performRateLimitedRequest(router, "10.0.0.1"); recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for exhausted client, got %d", recorder.Code)
	}
	if recorder := performRateLimitedRequest(router, "10.0.0.2"); recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for other client, got %d", recorder.Code)
	}
}

func TestRateLimiterRefillsOverTime(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	now := time.Now()

	if !limiter.allow("client", now) || !limiter.allow("client", now) {
		t.Fatal("expected initial burst to be allowed")
	}
	if limiter.allow("client", now) {
		t.Fatal("expected empty bucket to deny")
	}

	// Half the window refills half the capacity (one token).
	if !limiter.allow("client", now.Add(30*time.Second)) {
		t.Fatal("expected refilled token to be allowed")
	}
	if limiter.allow("client", now.Add(30*time.Second)) {
		t.Fatal("expected bucket to be empty again")
	}
}

func TestRateLimiterCleanupDropsIdleBuckets(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	now := time.Now()

	limiter.allow("idle-client", now)
	if len(limiter.buckets) != 1 {
		t.Fatalf("expected one bucket, got %d", len(limiter.buckets))
	}

	// After the window the idle bucket is back to capacity; the next
	// cleanup pass (once the cleanup interval elapsed) must drop it.
	limiter.allow("other-client", now.Add(2*time.Minute))
	if _, exists := limiter.buckets["idle-client"]; exists {
		t.Fatal("expected idle bucket to be cleaned up")
	}
}
