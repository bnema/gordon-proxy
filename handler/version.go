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

	// Extraire seulement la partie version du tag
	arm64Version, _, _ := parseVersion(arm64Tag.Tag.Name)
	amd64Version, _, _ := parseVersion(amd64Tag.Tag.Name)

	// Return the most recent version numbers for arm64 and amd64 as JSON
	return c.JSON(http.StatusOK, echo.Map{
		"arm64": arm64Version.String(),
		"amd64": amd64Version.String(),
	})
}

var releaseTagRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+-(arm64|amd64)$`)

func isValidReleaseTag(tag string) bool {
	return releaseTagRegex.MatchString(tag)
}

func getRecentVersions(metadata []ShortMetadata) (ShortMetadata, ShortMetadata, error) {
	logger := log.With().Str("func", "getRecentVersions").Logger()
	latestVersions := make(map[string]*semver.Version)
	latestMetadata := make(map[string]ShortMetadata)

	logger.Debug().Int("metadataCount", len(metadata)).Msg("Processing metadata")

	for _, m := range metadata {
		logger.Debug().Str("tag", m.Tag.Name).Msg("Processing tag")

		version, arch, err := parseVersion(m.Tag.Name)
		if err != nil {
			logger.Debug().Err(err).Str("tag", m.Tag.Name).Msg("Skipping invalid tag")
			continue
		}

		if latestVersions[arch] == nil || version.GreaterThan(latestVersions[arch]) {
			latestVersions[arch] = version
			latestMetadata[arch] = m
		}
	}

	logger.Debug().Interface("latestVersions", latestVersions).Msg("Processed versions")

	arm64Tag, arm64Ok := latestMetadata[ArchARM64]
	amd64Tag, amd64Ok := latestMetadata[ArchAMD64]

	if !arm64Ok || !amd64Ok {
		missingArchs := []string{}
		if !arm64Ok {
			missingArchs = append(missingArchs, ArchARM64)
		}
		if !amd64Ok {
			missingArchs = append(missingArchs, ArchAMD64)
		}
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("missing architectures: %v", missingArchs)
	}

	return arm64Tag, amd64Tag, nil
}

func parseVersion(tag string) (*semver.Version, string, error) {
	parts := strings.Split(tag, "-")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid tag format: %s", tag)
	}

	versionStr := strings.TrimPrefix(parts[0], "v")
	version, err := semver.NewVersion(versionStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse version: %w", err)
	}

	return version, parts[1], nil
}
