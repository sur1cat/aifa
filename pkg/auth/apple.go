package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidAppleToken = errors.New("invalid apple token")
)

type AppleUserInfo struct {
	Sub   string
	Email string
	Name  string
}

type AppleTokenVerifier struct {
	httpClient *http.Client
	keysCache  map[string]*rsa.PublicKey
	keysMux    sync.RWMutex
}

type appleKeysResponse struct {
	Keys []appleKey `json:"keys"`
}

type appleKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewAppleTokenVerifier() *AppleTokenVerifier {
	return &AppleTokenVerifier{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		keysCache: make(map[string]*rsa.PublicKey),
	}
}

func (v *AppleTokenVerifier) VerifyIDToken(ctx context.Context, idToken string) (*AppleUserInfo, error) {

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidAppleToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidAppleToken
	}

	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, ErrInvalidAppleToken
	}

	publicKey, err := v.getPublicKey(ctx, header.Kid)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidAppleToken
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, ErrInvalidAppleToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidAppleToken
	}

	iss, _ := claims["iss"].(string)
	if iss != "https://appleid.apple.com" {
		return nil, ErrInvalidAppleToken
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	if sub == "" {
		return nil, ErrInvalidAppleToken
	}

	return &AppleUserInfo{
		Sub:   sub,
		Email: email,
	}, nil
}

func (v *AppleTokenVerifier) getPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {

	v.keysMux.RLock()
	if key, ok := v.keysCache[kid]; ok {
		v.keysMux.RUnlock()
		return key, nil
	}
	v.keysMux.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keysResp appleKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&keysResp); err != nil {
		return nil, err
	}

	for _, key := range keysResp.Keys {
		if key.Kid == kid {
			publicKey, err := parseRSAPublicKey(key.N, key.E)
			if err != nil {
				return nil, err
			}

			v.keysMux.Lock()
			v.keysCache[kid] = publicKey
			v.keysMux.Unlock()

			return publicKey, nil
		}
	}

	return nil, ErrInvalidAppleToken
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

func ParseAppleUserName(userJSON string) string {
	if userJSON == "" {
		return ""
	}

	var user struct {
		Name struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"name"`
	}

	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return ""
	}

	name := strings.TrimSpace(fmt.Sprintf("%s %s", user.Name.FirstName, user.Name.LastName))
	return name
}
