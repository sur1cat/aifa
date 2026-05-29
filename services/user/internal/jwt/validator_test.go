package jwt

import (
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-secret-key-min-32-chars!!!!"

func signAccess(t *testing.T, secret string, uid uuid.UUID, ttl time.Duration) string {
	t.Helper()
	now := time.Now()
	c := Claims{
		UserID:    uid,
		TokenType: TokenAccess,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: gojwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  gojwt.NewNumericDate(now),
			Subject:   uid.String(),
		},
	}
	tok, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, c).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestValidateAccept(t *testing.T) {
	uid := uuid.New()
	tok := signAccess(t, testSecret, uid, time.Hour)
	claims, err := NewValidator(testSecret).ValidateAccess(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != uid {
		t.Fatalf("uid mismatch: %s vs %s", claims.UserID, uid)
	}
}

func TestRejectWrongSecret(t *testing.T) {
	tok := signAccess(t, testSecret, uuid.New(), time.Hour)
	if _, err := NewValidator("other-secret-min-32-chars!!!!!!!!").ValidateAccess(tok); err == nil {
		t.Fatal("accepted tampered token")
	}
}

func TestRejectExpired(t *testing.T) {
	tok := signAccess(t, testSecret, uuid.New(), -time.Minute)
	if _, err := NewValidator(testSecret).ValidateAccess(tok); err == nil {
		t.Fatal("accepted expired token")
	}
}

func TestRejectWrongType(t *testing.T) {
	now := time.Now()
	c := Claims{
		UserID:    uuid.New(),
		TokenType: "refresh",
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: gojwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  gojwt.NewNumericDate(now),
		},
	}
	tok, _ := gojwt.NewWithClaims(gojwt.SigningMethodHS256, c).SignedString([]byte(testSecret))
	if _, err := NewValidator(testSecret).ValidateAccess(tok); err == nil {
		t.Fatal("accepted refresh-typed token as access")
	}
}

func TestRejectGarbage(t *testing.T) {
	v := NewValidator(testSecret)
	for _, tc := range []string{"", "not-a-jwt", "a.b.c"} {
		if _, err := v.ValidateAccess(tc); err == nil {
			t.Fatalf("accepted %q", tc)
		}
	}
}
