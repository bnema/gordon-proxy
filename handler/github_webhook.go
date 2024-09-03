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
	"github.com/rs/zerolog/log"
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
	logger := log.With().Str("handler", "PostGithubWebhook").Logger()

	secret := client.WebhookSecret
	if secret == "" {
		logger.Error().Msg("Webhook secret is not set")
		return fmt.Errorf("webhook secret is not set")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read request body")
		return fmt.Errorf("failed to read request body: %w", err)
	}

	signatureSHA256 := c.Request().Header.Get("X-Hub-Signature-256")
	if signatureSHA256 == "" {
		logger.Warn().Msg("X-Hub-Signature-256 header is missing")
		return c.JSON(http.StatusUnauthorized, nil)
	}

	signature := strings.TrimPrefix(signatureSHA256, "sha256=")
	if signature == signatureSHA256 {
		logger.Warn().Msg("Signature prefix 'sha256=' is missing")
		return c.JSON(http.StatusBadRequest, nil)
	}

	computedSignature := generateSignature(secret, string(body))

	if !hmac.Equal([]byte(computedSignature), []byte(signature)) {
		logger.Warn().
			Str("received", signature).
			Str("computed", computedSignature).
			Msg("Signature mismatch")
		return c.JSON(http.StatusUnauthorized, nil)
	}

	bodyReader := bytes.NewReader(body)

	var event GitHubPackageEvent
	err = json.NewDecoder(bodyReader).Decode(&event)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to decode JSON payload")
		return fmt.Errorf("failed to decode JSON payload: %w", err)
	}

	metadata := event.Package.PackageVersion.ContainerMetadata
	err = SaveMetadataToFile(metadata)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to save metadata to file")
		return fmt.Errorf("failed to save metadata to file: %w", err)
	}

	logger.Info().Msg("Webhook processed successfully")
	return c.JSON(http.StatusOK, nil)
}

func generateSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
