package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a sliding window rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int           // max requests
	window   time.Duration // time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, timestamps := range rl.requests {
			// Filter out old timestamps
			var valid []time.Time
			for _, t := range timestamps {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}

// isAllowed checks if request is allowed and records it
func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timestamps := rl.requests[key]

	// Filter out old timestamps
	var valid []time.Time
	for _, t := range timestamps {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	// Check limit
	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	// Add current request
	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

// RateLimit creates a rate limiting middleware
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		// Use client IP as key
		key := c.ClientIP()

		// If user is authenticated, use user ID instead
		if userID, exists := c.Get("userID"); exists {
			key = userID.(interface{ String() string }).String()
		}

		if !limiter.isAllowed(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests, please try again later",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimit creates a stricter rate limiting for sensitive endpoints
func StrictRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		// Use client IP for strict limiting (prevents brute force across accounts)
		key := c.ClientIP()

		if !limiter.isAllowed(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests, please try again later",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
