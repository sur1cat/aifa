package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("access token is empty")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("refresh token is empty")
	}

	if tokenPair.AccessToken == tokenPair.RefreshToken {
		t.Error("access and refresh tokens should be different")
	}
}

func TestJWTManager_ValidateAccessToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	claims, err := jwtManager.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}

	if claims.TokenType != "access" {
		t.Errorf("expected token type 'access', got %s", claims.TokenType)
	}
}

func TestJWTManager_ValidateRefreshToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	claims, err := jwtManager.ValidateRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}

	if claims.TokenType != "refresh" {
		t.Errorf("expected token type 'refresh', got %s", claims.TokenType)
	}
}

func TestJWTManager_AccessTokenAsRefresh_ShouldFail(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	_, err = jwtManager.ValidateRefreshToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("expected error when using access token as refresh token")
	}
}

func TestJWTManager_RefreshTokenAsAccess_ShouldFail(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	_, err = jwtManager.ValidateAccessToken(tokenPair.RefreshToken)
	if err == nil {
		t.Error("expected error when using refresh token as access token")
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {

	jwtManager := NewJWTManager("test-secret-key-min-32-chars!!", -time.Hour, -time.Hour)
	userID := uuid.New()

	tokenPair, err := jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	_, err = jwtManager.ValidateAccessToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestJWTManager_InvalidSignature(t *testing.T) {
	jwtManager1 := NewJWTManager("secret-key-one-min-32-chars!!!!", time.Hour, 24*time.Hour)
	jwtManager2 := NewJWTManager("secret-key-two-min-32-chars!!!!", time.Hour, 24*time.Hour)
	userID := uuid.New()

	tokenPair, err := jwtManager1.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	_, err = jwtManager2.ValidateAccessToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key-min-32-chars!!", time.Hour, 24*time.Hour)

	testCases := []struct {
		name  string
		token string
	}{
		{"Empty token", ""},
		{"Invalid format", "not-a-jwt"},
		{"Malformed JWT", "header.payload.signature"},
		{"Random string", "abc123xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := jwtManager.ValidateAccessToken(tc.token)
			if err == nil {
				t.Error("expected error for invalid token")
			}
		})
	}
}
