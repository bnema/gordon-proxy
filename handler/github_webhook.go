package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Define a struct that matches the expected structure of your JSON payload
type GitHubWebhookEvent struct {
	Action string `json:"action"`
}

func PostGithubWebhook(c echo.Context, client *GitHubClient) error {
	secret := client.WebhookSecret
	if secret == "" {
		return fmt.Errorf("webhook secret is not set")
	}

	// Use the raw body for HMAC computation
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	signatureSHA256 := c.Request().Header.Get("X-Hub-Signature-256")
	if signatureSHA256 == "" {
		return fmt.Errorf("X-Hub-Signature-256 header is missing")
	}
	signature := strings.TrimPrefix(signatureSHA256, "sha256=")
	if signature == signatureSHA256 {
		return fmt.Errorf("X-Hub-Signature-256 header is malformed")
	}

	// Compute the HMAC
	computedSignature := generateSignature(secret, string(body))

	// Check if the computed HMAC matches the GitHub signature
	if !hmac.Equal([]byte(computedSignature), []byte(signature)) {
		return fmt.Errorf("invalid signature: computed %s, received %s", computedSignature, signature)
	}

	bodyReader := bytes.NewReader(body)

	// Decode the JSON payload into a map for inspection
	var payload map[string]interface{}
	err = json.NewDecoder(bodyReader).Decode(&payload)
	if err != nil {
		return fmt.Errorf("failed to decode JSON payload: %w", err)
	}

	// For debugging purposes, print the entire payload
	fmt.Printf("Received payload: %+v\n", payload)

	// Return 200 response to GitHub to acknowledge the webhook
	return c.JSON(http.StatusOK, nil)
}

func generateSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
