package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	ErrInvalidGoogleToken = errors.New("invalid google token")
)

type GoogleUserInfo struct {
	Sub           string      `json:"sub"`
	Email         string      `json:"email"`
	EmailVerified interface{} `json:"email_verified"`
	Name          string      `json:"name"`
	Picture       string      `json:"picture"`
}

func (g *GoogleUserInfo) IsEmailVerified() bool {
	switch v := g.EmailVerified.(type) {
	case bool:
		return v
	case string:
		return v == "true"
	default:
		return false
	}
}

type GoogleTokenVerifier struct {
	httpClient *http.Client
}

func NewGoogleTokenVerifier() *GoogleTokenVerifier {
	return &GoogleTokenVerifier{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (v *GoogleTokenVerifier) VerifyIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error) {
	log.Printf("Verifying Google token (length=%d)", len(idToken))

	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to verify token (network error): %v", err)
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Google API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {

		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		log.Printf("Google token verification failed: status=%d, response=%v", resp.StatusCode, errorResp)
		return nil, ErrInvalidGoogleToken
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Failed to decode Google response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Google user info: sub=%s, email=%s", userInfo.Sub, userInfo.Email)

	if userInfo.Sub == "" {
		log.Printf("Google user sub is empty")
		return nil, ErrInvalidGoogleToken
	}

	return &userInfo, nil
}
