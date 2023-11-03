package handler

import (
	"fmt"
	"net/http"
	"sort"
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

func getRecentVersions(metadata []ShortMetadata) (arm64Tag, amd64Tag ShortMetadata, err error) {
	// Filter out the 'latest' tag and separate arm64 and amd64
	var arm64Versions, amd64Versions []ShortMetadata
	for _, m := range metadata {
		if m.Tag.Name != "latest" {
			if strings.Contains(m.Tag.Name, "arm64") {
				arm64Versions = append(arm64Versions, m)
			} else if strings.Contains(m.Tag.Name, "amd64") {
				amd64Versions = append(amd64Versions, m)
			}
		}
	}

	versionSort := func(i, j int) bool {
		return strings.Split(arm64Versions[i].Tag.Name, "-")[0] > strings.Split(arm64Versions[j].Tag.Name, "-")[0]
	}

	// Sort and get the most recent versions
	sort.Slice(arm64Versions, versionSort)
	sort.Slice(amd64Versions, versionSort)

	if len(arm64Versions) > 0 {
		arm64Tag = arm64Versions[0] // The most recent version for arm64
	}
	if len(amd64Versions) > 0 {
		amd64Tag = amd64Versions[0] // The most recent version for amd64
	}

	return arm64Tag, amd64Tag, nil
}
