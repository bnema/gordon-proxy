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

	// Debug: Compare the computed signature with the expected signature
	fmt.Printf("Computed signature: %s\n", computedSignature)
	fmt.Printf("GitHub signature: %s\n", signature)

	// Check if the computed HMAC matches the GitHub signature
	if !hmac.Equal([]byte(computedSignature), []byte(signature)) {
		return fmt.Errorf("invalid signature: computed %s, received %s", computedSignature, signature)
	}

	return c.JSON(http.StatusOK, "Webhook processed successfully")
}

func generateSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
