package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

var currentVersion string

type GithubRelease struct {
	TagName string `json:"tag_name"`
}

// VersionService handles version checking and scheduling
type VersionService struct {
	client *http.Client
	ticker *time.Ticker
	done   chan bool
}

func NewVersionService() *VersionService {
	return &VersionService{
		client: &http.Client{Timeout: 10 * time.Second},
		done:   make(chan bool),
	}
}

func (vs *VersionService) fetchLatestVersion() error {
	resp, err := vs.client.Get("https://api.github.com/repos/bnema/gordon/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	currentVersion = release.TagName
	log.Info().Str("version", currentVersion).Msg("Updated latest version")
	return nil
}

func (vs *VersionService) Start() {
	// Initial fetch
	if err := vs.fetchLatestVersion(); err != nil {
		log.Error().Err(err).Msg("Failed to fetch initial version")
	}

	// Start periodic updates
	vs.ticker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-vs.ticker.C:
				if err := vs.fetchLatestVersion(); err != nil {
					log.Error().Err(err).Msg("Failed to fetch version update")
				}
			case <-vs.done:
				return
			}
		}
	}()
}

func (vs *VersionService) Stop() {
	if vs.ticker != nil {
		vs.ticker.Stop()
	}
	vs.done <- true
}

// GetVersion returns the current version
func GetVersion(c echo.Context) error {
	if currentVersion == "" {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{
			"error": "Version not yet available",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"version": currentVersion,
	})
}
