package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type GitHubWebHookHeaders struct {
	UserAgent                         string `header:"User-Agent"`
	XGitHubDelivery                   string `header:"X-GitHub-Delivery"`
	XGitHubEvent                      string `header:"X-GitHub-Event"`
	XGitHubHookID                     string `header:"X-GitHub-Hook-ID"`
	XGitHubHookInstallationTargetID   string `header:"X-GitHub-Hook-Installation-Target-ID"`
	XGitHubHookInstallationTargetType string `header:"X-GitHub-Hook-Installation-Target-Type"`
	XHubSignature                     string `header:"X-Hub-Signature"`
	XHubSignature256                  string `header:"X-Hub-Signature-256"`
}

type GitHubWebhookPayload interface{}

func PostGithubWebhook(c echo.Context, client *GitHubClient) error {
	headers := new(GitHubWebHookHeaders)
	if err := c.Bind(headers); err != nil {
		return err
	}

	// Read the body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return fmt.Errorf("error while reading request body: %w", err)
	}

	// Validate the signature (assuming SHA256 is used)
	expectedSignature := headers.XHubSignature256
	if expectedSignature == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "No X-Hub-Signature-256 header present in request")
	}

	// Compute HMAC with the secret
	mac := hmac.New(sha256.New, []byte(client.WebhookSecret))
	_, err = mac.Write(body)
	if err != nil {
		return fmt.Errorf("error while computing HMAC: %w", err)
	}
	computedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Compare the GitHub signature with your computed signature
	if !hmac.Equal([]byte(expectedSignature), []byte(computedSignature)) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid signature")
	}

	// Unmarshal payload into GitHubWebhookPayload interface
	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("error while unmarshalling payload: %w", err)
	}

	// Print or process the payload as needed
	fmt.Print(payload)

	// If the GitHub event is a 'ping', just respond with success
	if headers.XGitHubEvent == "ping" {
		return c.NoContent(http.StatusOK)
	}

	return c.NoContent(http.StatusOK)
}
