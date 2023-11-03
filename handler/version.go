package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Expose the /version endpoint
func GetLatestTags(c echo.Context) error {
	metadata, err := ReadShortMetadataFromFile()
	if err != nil {
		return fmt.Errorf("failed to read metadata from file: %w", err)
	}

	arm64Tag, amd64Tag, err := getRecentVersions(metadata)
	if err != nil {
		return err
	}

	// Return the most recent tags for arm64 and amd64 as JSON
	return c.JSON(http.StatusOK, echo.Map{
		"arm64": arm64Tag.Tag,
		"amd64": amd64Tag.Tag,
	})
}

func parseVersion(tag string) []int {
	// Assuming the tag format is "0.0.951-arm64" or "0.0.951-amd64"
	parts := strings.Split(tag, "-")
	if len(parts) != 2 {
		return nil
	}
	var versionParts []int
	for _, p := range strings.Split(parts[0], ".") {
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		versionParts = append(versionParts, v)
	}
	return versionParts
}

func compareVersions(v1, v2 []int) bool {
	for i := 0; i < len(v1) && i < len(v2); i++ {
		if v1[i] != v2[i] {
			return v1[i] > v2[i]
		}
	}
	return len(v1) > len(v2)
}

func getRecentVersions(metadata []ShortMetadata) (ShortMetadata, ShortMetadata, error) {
	// Create maps to hold the latest version for each architecture
	latestVersions := make(map[string]ShortMetadata)

	for _, m := range metadata {
		if m.Tag.Name == "latest" {
			continue // skip the 'latest' tag
		}
		tagParts := strings.Split(m.Tag.Name, "-")
		if len(tagParts) != 2 {
			continue
		}
		versionParts := parseVersion(m.Tag.Name)
		if versionParts == nil { // Skip if the version is not properly parsed
			continue
		}
		arch := tagParts[1]

		// Check if this is the latest version for the architecture
		if current, exists := latestVersions[arch]; !exists || (versionParts != nil && compareVersions(versionParts, parseVersion(current.Tag.Name))) {
			latestVersions[arch] = m
		}
	}

	arm64Tag, arm64Ok := latestVersions["arm64"]
	amd64Tag, amd64Ok := latestVersions["amd64"]

	if !arm64Ok || !amd64Ok {
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("one or both architectures are missing")
	}

	arm64Version := parseVersion(arm64Tag.Tag.Name)
	amd64Version := parseVersion(amd64Tag.Tag.Name)

	// If the versions are not the same length or any part of them does not match, return an error
	if len(arm64Version) != len(amd64Version) {
		return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("the latest versions length do not match for arm64 and amd64")
	}
	for i := range arm64Version {
		if arm64Version[i] != amd64Version[i] {
			return ShortMetadata{}, ShortMetadata{}, fmt.Errorf("the latest versions do not match for arm64 and amd64")
		}
	}

	return arm64Tag, amd64Tag, nil
}
