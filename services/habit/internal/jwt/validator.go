package jwt

import (
	"errors"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalid = errors.New("invalid token")
	ErrExpired = errors.New("token expired")
)

type TokenType string

const TokenAccess TokenType = "access"

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	gojwt.RegisteredClaims
}

type Validator struct {
	secret []byte
}

func NewValidator(secret string) *Validator {
	return &Validator{secret: []byte(secret)}
}

func (v *Validator) ValidateAccess(token string) (*Claims, error) {
	parsed, err := gojwt.ParseWithClaims(token, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalid
		}
		return v.secret, nil
	})
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || claims.TokenType != TokenAccess {
		return nil, ErrInvalid
	}
	return claims, nil
}
