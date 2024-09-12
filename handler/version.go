package handler

import (
	"encoding/json"
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

	logger.Debug().Int("metadataCount", len(metadata)).Msg("Processing metadata")

	for _, m := range metadata {
		logger.Debug().Str("tag", m.Tag.Name).Msg("Processing tag")

		if m.Tag.Name == "latest" {
			// Instead of skipping, let's try to extract architecture information from the labels
			if platforms, ok := m.Labels.AllLabels["github.internal.platforms"]; ok {
				var platformInfo []struct {
					Digest       string `json:"digest"`
					Architecture string `json:"architecture"`
					OS           string `json:"os"`
				}
				if err := json.Unmarshal([]byte(platforms), &platformInfo); err != nil {
					logger.Warn().Err(err).Msg("Failed to parse platform info")
					continue
				}

				for _, p := range platformInfo {
					latestVersions[p.Architecture] = ShortMetadata{
						Tag: struct {
							Name   string `json:"name"`
							Digest string `json:"digest"`
						}{
							Name:   fmt.Sprintf("latest-%s", p.Architecture),
							Digest: p.Digest,
						},
					}
				}
			}
			continue
		}

		// ... (rest of the existing logic for processing other tags)
	}

	logger.Debug().Interface("latestVersions", latestVersions).Msg("Processed versions")

	arm64Tag, arm64Ok := latestVersions["arm64"]
	amd64Tag, amd64Ok := latestVersions["amd64"]

	if !arm64Ok || !amd64Ok {
		missingArchs := []string{}
		if !arm64Ok {
			missingArchs = append(missingArchs, "arm64")
		}
		if !amd64Ok {
			missingArchs = append(missingArchs, "amd64")
		}
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("missing architectures: %v", missingArchs)
	}

	return arm64Tag, amd64Tag, nil
}
