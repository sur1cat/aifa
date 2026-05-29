package apns

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

const (
	tokenTTL    = 50 * time.Minute
	hostProd    = "https://api.push.apple.com"
	hostSandbox = "https://api.sandbox.push.apple.com"
)

var ErrNotConfigured = errors.New("APNS is not configured")

type Notification struct {
	DeviceToken string
	Title       string
	Body        string
	Badge       *int
	Sound       string
	Data        map[string]any
}

// Client signs ephemeral provider tokens (50-min TTL) and posts to APNs over
// HTTP/2. Caches the signed JWT to avoid re-signing on every call.
type Client struct {
	keyID, teamID, bundleID string
	privateKey              *ecdsa.PrivateKey
	host                    string

	tokenMu   sync.Mutex
	token     string
	tokenTime time.Time

	http *http.Client
}

func NewClient(keyPath, keyID, teamID, bundleID string, production bool) (*Client, error) {
	if keyPath == "" || keyID == "" || teamID == "" || bundleID == "" {
		return nil, ErrNotConfigured
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read APNS key: %w", err)
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, errors.New("APNS key: failed to parse PEM block")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 key: %w", err)
	}
	pk, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("APNS key is not ECDSA")
	}

	host := hostSandbox
	if production {
		host = hostProd
	}

	return &Client{
		keyID:      keyID,
		teamID:     teamID,
		bundleID:   bundleID,
		privateKey: pk,
		host:       host,
		http:       &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) providerToken() (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.token != "" && time.Since(c.tokenTime) < tokenTTL {
		return c.token, nil
	}

	now := time.Now()
	t := gojwt.NewWithClaims(gojwt.SigningMethodES256, gojwt.MapClaims{
		"iss": c.teamID,
		"iat": now.Unix(),
	})
	t.Header["kid"] = c.keyID

	signed, err := t.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign provider token: %w", err)
	}
	c.token = signed
	c.tokenTime = now
	return signed, nil
}

type payload struct {
	Aps  apsBody        `json:"aps"`
	Data map[string]any `json:"data,omitempty"`
}

type apsBody struct {
	Alert apsAlert `json:"alert"`
	Badge *int     `json:"badge,omitempty"`
	Sound string   `json:"sound,omitempty"`
}

type apsAlert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (c *Client) Send(ctx context.Context, n *Notification) error {
	tok, err := c.providerToken()
	if err != nil {
		return err
	}

	sound := n.Sound
	if sound == "" {
		sound = "default"
	}
	body, err := json.Marshal(payload{
		Aps:  apsBody{Alert: apsAlert{Title: n.Title, Body: n.Body}, Badge: n.Badge, Sound: sound},
		Data: n.Data,
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/3/device/%s", c.host, n.DeviceToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "bearer "+tok)
	req.Header.Set("apns-topic", c.bundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("APNs status=%d body=%s", resp.StatusCode, string(raw))
	}
	return nil
}
