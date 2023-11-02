package handler

import (
	"crypto/hmac"
	"crypto/sha1"
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

func BindGithubWebhookRoute(e *echo.Echo, client *GitHubClient) {
	e.POST("/github-proxy/webhooks", func(c echo.Context) error {
		headers := new(GitHubWebHookHeaders)
		if err := c.Bind(headers); err != nil {
			return err
		}

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		// Check X-Hub-Signature
		if !verifySignature(client.WebhookSecret, headers.XHubSignature, body) {
			return c.NoContent(http.StatusUnauthorized)
		}

		// Check X-Hub-Signature-256
		if !verifySignature(client.WebhookSecret, headers.XHubSignature256, body) {
			return c.NoContent(http.StatusUnauthorized)
		}

		// Unmarshal payload into GitHubWebhookPayload interface because we don't know the type yet
		var payload GitHubWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		fmt.Print(payload)

		return c.NoContent(http.StatusOK)
	})
}

func verifySignature(secret, signature string, payload []byte) bool {
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte("sha1="+expectedMAC), []byte(signature))
}
