package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, timestamps := range rl.requests {

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

func (rl *RateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timestamps := rl.requests[key]

	var valid []time.Time
	for _, t := range timestamps {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {

		key := c.ClientIP()

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

func StrictRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {

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
