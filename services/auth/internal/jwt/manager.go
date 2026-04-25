package jwt

import (
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalid = errors.New("invalid token")
	ErrExpired = errors.New("token expired")
)

type TokenType string

const (
	TokenAccess  TokenType = "access"
	TokenRefresh TokenType = "refresh"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	gojwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *Manager) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	access, err := m.sign(userID, TokenAccess, now, now.Add(m.accessTTL))
	if err != nil {
		return nil, err
	}
	refresh, err := m.sign(userID, TokenRefresh, now, now.Add(m.refreshTTL))
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresAt: now.Add(m.accessTTL).Unix()}, nil
}

func (m *Manager) sign(userID uuid.UUID, typ TokenType, now, exp time.Time) (string, error) {
	claims := Claims{
		UserID:    userID,
		TokenType: typ,
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: gojwt.NewNumericDate(exp),
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			Issuer:    "aifa",
			Subject:   userID.String(),
		},
	}
	return gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) ValidateAccess(token string) (*Claims, error) {
	return m.validate(token, TokenAccess)
}

func (m *Manager) ValidateRefresh(token string) (*Claims, error) {
	return m.validate(token, TokenRefresh)
}

func (m *Manager) validate(token string, expected TokenType) (*Claims, error) {
	parsed, err := gojwt.ParseWithClaims(token, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || claims.TokenType != expected {
		return nil, ErrInvalid
	}
	return claims, nil
}

func (m *Manager) RefreshAccess(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateRefresh(refreshToken)
	if err != nil {
		return nil, err
	}
	return m.GenerateTokenPair(claims.UserID)
}
