package handler

import (
	"fmt"
	"net/http"
	"regexp"
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

var releaseTagRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+-(arm64|amd64)$`)

func isValidReleaseTag(tag string) bool {
	return releaseTagRegex.MatchString(tag)
}

func parseVersion(tag string) (*semver.Version, string, error) {
	if !isValidReleaseTag(tag) {
		return nil, "", fmt.Errorf("invalid release tag format: %s", tag)
	}

	parts := strings.Split(tag, "-")
	versionStr := strings.TrimPrefix(parts[0], "v")
	version, err := semver.NewVersion(versionStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse version: %w", err)
	}

	return version, parts[1], nil
}

func getRecentVersions(metadata []ShortMetadata) (ShortMetadata, ShortMetadata, error) {
	logger := log.With().Str("func", "getRecentVersions").Logger()
	latestVersions := make(map[string]ShortMetadata)

	for _, m := range metadata {
		if m.Tag.Name == "latest" {
			logger.Debug().Str("tag", m.Tag.Name).Msg("Skipping 'latest' tag")
			continue
		}

		version, arch, err := parseVersion(m.Tag.Name)
		if err != nil {
			logger.Warn().Err(err).Str("tag", m.Tag.Name).Msg("Skipping invalid tag")
			continue
		}

		if current, exists := latestVersions[arch]; !exists {
			latestVersions[arch] = m
			logger.Debug().Str("arch", arch).Str("version", version.String()).Msg("Set initial version for architecture")
		} else {
			currentVersion, _, _ := parseVersion(current.Tag.Name)
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

	arm64Version, _, err := parseVersion(arm64Tag.Tag.Name)
	if err != nil {
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("failed to parse arm64 version: %w", err)
	}

	amd64Version, _, err := parseVersion(amd64Tag.Tag.Name)
	if err != nil {
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("failed to parse amd64 version: %w", err)
	}

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
