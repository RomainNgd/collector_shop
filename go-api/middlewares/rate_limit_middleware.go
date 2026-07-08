package middlewares

import (
	"net/http"
	"poc-gin/controllers"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const rateLimiterCleanupInterval = time.Minute

type rateLimiterBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter is an in-memory token bucket keyed by client IP. Each key gets
// `maxRequests` tokens refilled linearly over `window`. It protects a single
// instance; a shared store (e.g. Redis) would be needed once the API scales
// horizontally.
type RateLimiter struct {
	mu           sync.Mutex
	buckets      map[string]*rateLimiterBucket
	capacity     float64
	refillPerSec float64
	window       time.Duration
	lastCleanup  time.Time
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	if maxRequests < 1 {
		maxRequests = 1
	}
	if window <= 0 {
		window = time.Minute
	}

	return &RateLimiter{
		buckets:      make(map[string]*rateLimiterBucket),
		capacity:     float64(maxRequests),
		refillPerSec: float64(maxRequests) / window.Seconds(),
		window:       window,
		lastCleanup:  time.Now(),
	}
}

func (rl *RateLimiter) allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.cleanupLocked(now)

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &rateLimiterBucket{tokens: rl.capacity, lastRefill: now}
		rl.buckets[key] = bucket
	} else {
		elapsed := now.Sub(bucket.lastRefill).Seconds()
		if elapsed > 0 {
			bucket.tokens = min(rl.capacity, bucket.tokens+elapsed*rl.refillPerSec)
			bucket.lastRefill = now
		}
	}

	if bucket.tokens < 1 {
		return false
	}

	bucket.tokens--
	return true
}

// cleanupLocked drops buckets that are back to full capacity so idle clients
// do not grow the map forever. Callers must hold rl.mu.
func (rl *RateLimiter) cleanupLocked(now time.Time) {
	if now.Sub(rl.lastCleanup) < rateLimiterCleanupInterval {
		return
	}
	rl.lastCleanup = now

	for key, bucket := range rl.buckets {
		elapsed := now.Sub(bucket.lastRefill).Seconds()
		if bucket.tokens+elapsed*rl.refillPerSec >= rl.capacity {
			delete(rl.buckets, key)
		}
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	retryAfter := strconv.Itoa(int(rl.window.Seconds()))

	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP(), time.Now()) {
			c.Header("Retry-After", retryAfter)
			controllers.RespondError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests, please retry later", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
