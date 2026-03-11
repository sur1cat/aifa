package push

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type APNsClient struct {
	keyID      string
	teamID     string
	bundleID   string
	privateKey *ecdsa.PrivateKey
	token      string
	tokenMux   sync.RWMutex
	tokenTime  time.Time
	production bool
	client     *http.Client
}

type Notification struct {
	DeviceToken string
	Title       string
	Body        string
	Badge       *int
	Sound       string
	Data        map[string]interface{}
}

type APNsPayload struct {
	Aps  APNsAps                `json:"aps"`
	Data map[string]interface{} `json:"data,omitempty"`
}

type APNsAps struct {
	Alert APNsAlert `json:"alert"`
	Badge *int      `json:"badge,omitempty"`
	Sound string    `json:"sound,omitempty"`
}

type APNsAlert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewAPNsClient(keyPath, keyID, teamID, bundleID string, production bool) (*APNsClient, error) {

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an ECDSA private key")
	}

	return &APNsClient{
		keyID:      keyID,
		teamID:     teamID,
		bundleID:   bundleID,
		privateKey: ecdsaKey,
		production: production,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *APNsClient) generateToken() (string, error) {
	c.tokenMux.RLock()
	if c.token != "" && time.Since(c.tokenTime) < 50*time.Minute {
		defer c.tokenMux.RUnlock()
		return c.token, nil
	}
	c.tokenMux.RUnlock()

	c.tokenMux.Lock()
	defer c.tokenMux.Unlock()

	if c.token != "" && time.Since(c.tokenTime) < 50*time.Minute {
		return c.token, nil
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": c.teamID,
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID

	signedToken, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	c.token = signedToken
	c.tokenTime = now

	return signedToken, nil
}

func (c *APNsClient) Send(notification *Notification) error {
	token, err := c.generateToken()
	if err != nil {
		return err
	}

	payload := APNsPayload{
		Aps: APNsAps{
			Alert: APNsAlert{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Badge: notification.Badge,
			Sound: notification.Sound,
		},
		Data: notification.Data,
	}

	if payload.Aps.Sound == "" {
		payload.Aps.Sound = "default"
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var host string
	if c.production {
		host = "https://api.push.apple.com"
	} else {
		host = "https://api.sandbox.push.apple.com"
	}

	url := fmt.Sprintf("%s/3/device/%s", host, notification.DeviceToken)

	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "bearer "+token)
	req.Header.Set("apns-topic", c.bundleID)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("APNs error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *APNsClient) SendToMany(tokens []string, title, body string, data map[string]interface{}) []error {
	var errors []error

	for _, token := range tokens {
		err := c.Send(&Notification{
			DeviceToken: token,
			Title:       title,
			Body:        body,
			Data:        data,
		})
		if err != nil {
			errors = append(errors, fmt.Errorf("token %s: %w", token[:10]+"...", err))
		}
	}

	return errors
}
