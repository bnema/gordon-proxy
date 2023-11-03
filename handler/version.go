package handler

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
)

// Expose the /version endpoint
func GetLatestTag(c echo.Context) error {
	metadata, err := ReadShortMetadataFromFile()
	if err != nil {
		return fmt.Errorf("failed to read metadata from file: %w", err)
	}

	// Assuming the tags can be sorted lexically to find the most recent version
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Tag.Name > metadata[j].Tag.Name
	})

	// Return the most recent tag as JSON
	if len(metadata) > 0 {
		return c.JSON(http.StatusOK, metadata[0].Tag)
	}

	return c.JSON(http.StatusNotFound, echo.Map{"message": "No metadata found"})
}
