package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-secret-key-min-32-chars!!!!"

func TestGenerateAndValidate(t *testing.T) {
	m := NewManager(testSecret, time.Hour, 24*time.Hour)
	uid := uuid.New()

	pair, err := m.GenerateTokenPair(uid)
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("empty tokens")
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Fatal("tokens must differ")
	}

	claims, err := m.ValidateAccess(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != uid {
		t.Fatalf("uid mismatch: %s vs %s", claims.UserID, uid)
	}
	if claims.ID == "" {
		t.Fatal("jti must be set")
	}
}

func TestTokenTypeMismatch(t *testing.T) {
	m := NewManager(testSecret, time.Hour, 24*time.Hour)
	pair, _ := m.GenerateTokenPair(uuid.New())

	if _, err := m.ValidateAccess(pair.RefreshToken); err == nil {
		t.Fatal("refresh accepted as access")
	}
	if _, err := m.ValidateRefresh(pair.AccessToken); err == nil {
		t.Fatal("access accepted as refresh")
	}
}

func TestExpired(t *testing.T) {
	m := NewManager(testSecret, -time.Hour, -time.Hour)
	pair, _ := m.GenerateTokenPair(uuid.New())
	if _, err := m.ValidateAccess(pair.AccessToken); err == nil {
		t.Fatal("expired token accepted")
	}
}

func TestTamperedSignature(t *testing.T) {
	m1 := NewManager("secret-one-min-32-chars!!!!!!!!!", time.Hour, time.Hour)
	m2 := NewManager("secret-two-min-32-chars!!!!!!!!!", time.Hour, time.Hour)
	pair, _ := m1.GenerateTokenPair(uuid.New())
	if _, err := m2.ValidateAccess(pair.AccessToken); err == nil {
		t.Fatal("token from other secret accepted")
	}
}

func TestInvalidInputs(t *testing.T) {
	m := NewManager(testSecret, time.Hour, time.Hour)
	cases := []string{"", "not-a-jwt", "a.b.c"}
	for _, tc := range cases {
		if _, err := m.ValidateAccess(tc); err == nil {
			t.Fatalf("accepted %q", tc)
		}
	}
}
