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

type GitHubPackageEvent struct {
	Package struct {
		PackageVersion struct {
			ContainerMetadata ContainerMetadata `json:"container_metadata"`
		} `json:"package_version"`
	} `json:"package"`
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
		return c.JSON(http.StatusUnauthorized, nil)
	}
	signature := strings.TrimPrefix(signatureSHA256, "sha256=")
	if signature == signatureSHA256 {
		return c.JSON(http.StatusBadRequest, nil)
	}

	// Compute the HMAC
	computedSignature := generateSignature(secret, string(body))

	// Check if the computed HMAC matches the GitHub signature
	if !hmac.Equal([]byte(computedSignature), []byte(signature)) {
		return c.JSON(http.StatusUnauthorized, nil)
	}

	bodyReader := bytes.NewReader(body)

	// Decode the JSON payload into the PackageEvent struct
	var event GitHubPackageEvent
	err = json.NewDecoder(bodyReader).Decode(&event)
	if err != nil {
		return fmt.Errorf("failed to decode JSON payload: %w", err)
	}
	metadata := event.Package.PackageVersion.ContainerMetadata
	// Save the metadata to the file
	err = SaveMetadataToFile(metadata)
	if err != nil {
		return fmt.Errorf("failed to save metadata to file: %w", err)
	}
	// Return 200 response to GitHub to acknowledge the webhook
	return c.JSON(http.StatusOK, nil)
}

func generateSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
