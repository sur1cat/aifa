package oauth

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

	gojwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"
)

var ErrInvalidAppleToken = errors.New("invalid apple token")

type AppleUserInfo struct {
	Sub   string
	Email string
	Name  string
}

type AppleVerifier struct {
	client   *http.Client
	clientID string
	keys     map[string]*rsa.PublicKey
	keysMu   sync.RWMutex
	group    singleflight.Group
}

func NewAppleVerifier(clientID string) *AppleVerifier {
	return &AppleVerifier{
		client:   &http.Client{Timeout: 10 * time.Second},
		clientID: clientID,
		keys:     make(map[string]*rsa.PublicKey),
	}
}

func (v *AppleVerifier) Verify(ctx context.Context, idToken string) (*AppleUserInfo, error) {
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
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, ErrInvalidAppleToken
	}

	pub, err := v.publicKey(ctx, header.Kid)
	if err != nil {
		return nil, err
	}

	token, err := gojwt.Parse(idToken, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidAppleToken
		}
		return pub, nil
	})
	if err != nil {
		return nil, ErrInvalidAppleToken
	}

	claims, ok := token.Claims.(gojwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidAppleToken
	}
	if iss, _ := claims["iss"].(string); iss != "https://appleid.apple.com" {
		return nil, ErrInvalidAppleToken
	}
	if v.clientID != "" {
		if aud, _ := claims["aud"].(string); aud != v.clientID {
			return nil, ErrInvalidAppleToken
		}
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, ErrInvalidAppleToken
	}
	email, _ := claims["email"].(string)
	return &AppleUserInfo{Sub: sub, Email: email}, nil
}

func (v *AppleVerifier) publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.keysMu.RLock()
	if key, ok := v.keys[kid]; ok {
		v.keysMu.RUnlock()
		return key, nil
	}
	v.keysMu.RUnlock()

	// Dedupe concurrent misses for the same kid so one HTTP fetch serves all callers.
	result, err, _ := v.group.Do(kid, func() (any, error) {
		return v.fetchKey(ctx, kid)
	})
	if err != nil {
		return nil, err
	}
	return result.(*rsa.PublicKey), nil
}

func (v *AppleVerifier) fetchKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	var match *rsa.PublicKey
	v.keysMu.Lock()
	for _, k := range body.Keys {
		pub, err := parseRSA(k.N, k.E)
		if err != nil {
			continue
		}
		v.keys[k.Kid] = pub
		if k.Kid == kid {
			match = pub
		}
	}
	v.keysMu.Unlock()

	if match == nil {
		return nil, fmt.Errorf("%w: unknown kid %q", ErrInvalidAppleToken, kid)
	}
	return match, nil
}

func parseRSA(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}

func ParseAppleUserName(payload string) string {
	if payload == "" {
		return ""
	}
	var u struct {
		Name struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"name"`
	}
	if err := json.Unmarshal([]byte(payload), &u); err != nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", u.Name.FirstName, u.Name.LastName))
}
