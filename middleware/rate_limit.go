package middleware

import (
	"net/http"
	"sync"
	"time"

	"otp-auth-backend/config"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for different clients
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

// getLimiter returns or creates a rate limiter for a specific key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// cleanup removes old limiters to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Simple cleanup - in production you might want more sophisticated cleanup
	if len(rl.limiters) > 10000 {
		rl.limiters = make(map[string]*rate.Limiter)
	}
}

// RateLimitMiddleware provides rate limiting functionality
func RateLimitMiddleware(config *config.Config) gin.HandlerFunc {
	if !config.Security.EnableRateLimit {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	// Calculate rate limit from config
	requestsPerSecond := float64(config.RateLimit.MaxRequests) / config.RateLimit.Window.Seconds()
	rateLimit := rate.Limit(requestsPerSecond)
	burst := config.RateLimit.MaxRequests

	limiter := NewRateLimiter(rateLimit, burst)

	// Cleanup old limiters periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return gin.HandlerFunc(func(c *gin.Context) {
		// Use client IP as rate limit key (or user ID if authenticated)
		key := c.ClientIP()

		if !limiter.getLimiter(key).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests",
				"retry_after": time.Now().Add(config.RateLimit.Window).Unix(),
				"limit":       config.RateLimit.MaxRequests,
				"window":      config.RateLimit.Window.String(),
			})
			c.Abort()
			return
		}

		c.Next()
	})
}
