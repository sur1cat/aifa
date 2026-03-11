package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"habitflow/pkg/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	authMiddleware := NewAuthMiddleware(jwtManager)

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	authMiddleware := NewAuthMiddleware(jwtManager)

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	authMiddleware := NewAuthMiddleware(jwtManager)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {

		ctxUserID, exists := c.Get("userID")
		if !exists {
			t.Error("userID not set in context")
		}
		if ctxUserID.(uuid.UUID) != userID {
			t.Errorf("expected userID %s, got %s", userID, ctxUserID)
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {

	jwtManager := auth.NewJWTManager("test-secret-key-min-32-chars!!", -time.Hour, 24*time.Hour)
	authMiddleware := NewAuthMiddleware(jwtManager)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for expired token, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	authMiddleware := NewAuthMiddleware(jwtManager)

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "some-token"},
		{"Empty Bearer", "Bearer "},
		{"Just Bearer", "Bearer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tc.header)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("UserID exists", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		expectedID := uuid.New()
		c.Set("userID", expectedID)

		id, ok := GetUserID(c)
		if !ok {
			t.Error("expected GetUserID to return true")
		}
		if id != expectedID {
			t.Errorf("expected %s, got %s", expectedID, id)
		}
	})

	t.Run("UserID does not exist", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		_, ok := GetUserID(c)
		if ok {
			t.Error("expected GetUserID to return false")
		}
	})
}
