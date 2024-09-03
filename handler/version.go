package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

const (
	ArchARM64 = "arm64"
	ArchAMD64 = "amd64"
)

// GetLatestTags exposes the /version endpoint
func GetLatestTags(c echo.Context) error {
	logger := log.With().Str("handler", "GetLatestTags").Logger()

	metadata, err := ReadShortMetadataFromFile()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read metadata from file")
		return fmt.Errorf("failed to read metadata from file: %w", err)
	}

	arm64Tag, amd64Tag, err := getRecentVersions(metadata)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get recent versions")
		return err
	}

	logger.Info().
		Str("arm64", arm64Tag.Tag.Name).
		Str("amd64", amd64Tag.Tag.Name).
		Msg("Retrieved latest tags")

	// Return the most recent tags for arm64 and amd64 as JSON
	return c.JSON(http.StatusOK, echo.Map{
		"arm64": arm64Tag.Tag.Name,
		"amd64": amd64Tag.Tag.Name,
	})
}

func parseVersion(tag string) (*semver.Version, error) {
	parts := strings.Split(tag, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid tag format: %s", tag)
	}
	return semver.NewVersion(parts[0])
}

func getRecentVersions(metadata []ShortMetadata) (ShortMetadata, ShortMetadata, error) {
	logger := log.With().Str("func", "getRecentVersions").Logger()
	latestVersions := make(map[string]ShortMetadata)

	for _, m := range metadata {
		if m.Tag.Name == "latest" {
			logger.Debug().Str("tag", m.Tag.Name).Msg("Skipping 'latest' tag")
			continue
		}

		version, err := parseVersion(m.Tag.Name)
		if err != nil {
			logger.Warn().Err(err).Str("tag", m.Tag.Name).Msg("Failed to parse version")
			continue
		}

		parts := strings.Split(m.Tag.Name, "-")
		if len(parts) != 2 {
			logger.Warn().Str("tag", m.Tag.Name).Msg("Invalid tag format")
			continue
		}
		arch := parts[1]

		if current, exists := latestVersions[arch]; !exists {
			latestVersions[arch] = m
			logger.Debug().Str("arch", arch).Str("version", version.String()).Msg("Set initial version for architecture")
		} else {
			currentVersion, _ := parseVersion(current.Tag.Name)
			if version.GreaterThan(currentVersion) {
				latestVersions[arch] = m
				logger.Debug().
					Str("arch", arch).
					Str("oldVersion", currentVersion.String()).
					Str("newVersion", version.String()).
					Msg("Updated to newer version for architecture")
			}
		}
	}

	arm64Tag, arm64Ok := latestVersions[ArchARM64]
	amd64Tag, amd64Ok := latestVersions[ArchAMD64]

	if !arm64Ok || !amd64Ok {
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("one or both architectures are missing")
	}

	arm64Version, _ := parseVersion(arm64Tag.Tag.Name)
	amd64Version, _ := parseVersion(amd64Tag.Tag.Name)

	if !arm64Version.Equal(amd64Version) {
		logger.Warn().
			Str("arm64", arm64Version.String()).
			Str("amd64", amd64Version.String()).
			Msg("Latest versions do not match for arm64 and amd64")
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("the latest versions do not match for arm64 and amd64")
	}

	logger.Info().
		Str("version", arm64Version.String()).
		Msg("Found matching latest version for both architectures")

	return arm64Tag, amd64Tag, nil
}
