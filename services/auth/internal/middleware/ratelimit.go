package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipBucket struct {
	tokens    int
	resetAt   time.Time
}

// RateLimiter ограничивает число запросов с одного IP.
// Алгоритм: фиксированное окно (fixed window) — maxRequests запросов за windowDur.
type RateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*ipBucket
	maxRequests int
	windowDur   time.Duration
}

func NewRateLimiter(maxRequests int, windowDur time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:     make(map[string]*ipBucket),
		maxRequests: maxRequests,
		windowDur:   windowDur,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, b := range rl.buckets {
			if now.After(b.resetAt) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[ip]
	if !ok || now.After(b.resetAt) {
		rl.buckets[ip] = &ipBucket{tokens: rl.maxRequests - 1, resetAt: now.Add(rl.windowDur)}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// Limit возвращает gin.HandlerFunc, блокирующий избыточные запросы с одного IP.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.Header("Retry-After", rl.windowDur.String())
			abort(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
				"Too many requests — please slow down")
			return
		}
		c.Next()
	}
}
