package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type GitHubWebhookPayload interface{}

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

	// Parse the payload
	var payload GitHubWebhookPayload
	err = c.Bind(&payload)
	if err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	fmt.Printf("Received webhook payload: %+v\n", payload)

	// Return 200 response to GitHub to acknowledge the webhook
	return c.JSON(http.StatusOK, nil)
}

func generateSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
