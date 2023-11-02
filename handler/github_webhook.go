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

type PackageEvent struct {
	ContainerMetadata ContainerMetadata `json:"container_metadata"`
}

type ContainerMetadata struct {
	Tag      Tag      `json:"tag"`
	Labels   Labels   `json:"labels"`
	Manifest Manifest `json:"manifest"`
}

type Tag struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

type Labels struct {
	Description string            `json:"description"`
	Source      string            `json:"source"`
	Revision    string            `json:"revision"`
	ImageURL    string            `json:"image_url"`
	Licenses    string            `json:"licenses"`
	AllLabels   map[string]string `json:"all_labels"`
}

type Manifest struct {
	Digest    string  `json:"digest"`
	MediaType string  `json:"media_type"`
	URI       string  `json:"uri"`
	Size      int     `json:"size"`
	Config    Config  `json:"config"`
	Layers    []Layer `json:"layers"`
}

type Config struct {
	Digest    string `json:"digest"`
	MediaType string `json:"media_type"`
	Size      int    `json:"size"`
}

type Layer struct {
	Digest    string `json:"digest"`
	MediaType string `json:"media_type"`
	Size      int    `json:"size"`
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

	// Decode the JSON payload into the PackageEvent struct
	var event PackageEvent
	err = json.NewDecoder(bodyReader).Decode(&event)
	if err != nil {
		return fmt.Errorf("failed to decode JSON payload: %w", err)
	}

	fmt.Printf("Tag name: %s\n", event.ContainerMetadata.Tag.Name)
	fmt.Printf("Tag digest: %s\n", event.ContainerMetadata.Tag.Digest)
	fmt.Printf("Source URL: %s\n", event.ContainerMetadata.Labels.AllLabels["org.opencontainers.image.source"])
	fmt.Printf("Image version: %s\n", event.ContainerMetadata.Labels.AllLabels["org.opencontainers.image.version"])
	fmt.Printf("Image revision: %s\n", event.ContainerMetadata.Labels.AllLabels["org.opencontainers.image.revision"])

	// Return 200 response to GitHub to acknowledge the webhook
	return c.JSON(http.StatusOK, nil)
}

func generateSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
