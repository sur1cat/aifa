package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimit_AllowsRequestsUnderLimit(t *testing.T) {
	router := gin.New()
	router.Use(RateLimit(5, time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	userID := uuid.New()

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	_ = userID
}

func TestRateLimit_BlocksExcessRequests(t *testing.T) {
	router := gin.New()
	router.Use(RateLimit(2, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	var blockedCount int
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			blockedCount++
		}
	}

	if blockedCount == 0 {
		t.Error("expected some requests to be rate limited")
	}
}

func TestStrictRateLimit_UsesIPOnly(t *testing.T) {
	router := gin.New()
	router.Use(StrictRateLimit(2, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected rate limit, got status %d", w.Code)
	}
}

func TestRateLimit_DifferentUsersHaveSeparateLimits(t *testing.T) {
	router := gin.New()
	router.Use(StrictRateLimit(1, time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("user1 first request: expected %d, got %d", http.StatusOK, w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("user2 first request: expected %d, got %d", http.StatusOK, w2.Code)
	}
}
