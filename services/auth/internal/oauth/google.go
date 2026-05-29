package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

var ErrInvalidGoogleToken = errors.New("invalid google token")

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Aud           string `json:"aud"`
}

type GoogleVerifier struct {
	client   *http.Client
	clientID string
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{
		client:   &http.Client{Timeout: 10 * time.Second},
		clientID: clientID,
	}
}

func (v *GoogleVerifier) Verify(ctx context.Context, idToken string) (*GoogleUserInfo, error) {
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tokeninfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		slog.Warn("google tokeninfo non-200", "status", resp.StatusCode, "body", body)
		return nil, ErrInvalidGoogleToken
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if info.Sub == "" {
		return nil, ErrInvalidGoogleToken
	}
	if v.clientID != "" && info.Aud != "" && info.Aud != v.clientID {
		return nil, ErrInvalidGoogleToken
	}
	return &info, nil
}
