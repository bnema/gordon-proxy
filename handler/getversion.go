package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
)

var mutex sync.RWMutex

// The file where the  metadata will be stored
const MetadataFilePath = "./metadata.json"

func SaveMetadataToFile(metadata ContainerMetadata) error {
	mutex.Lock()
	defer mutex.Unlock()

	// Marshal the metadata into JSON
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Write the JSON data to the file
	return os.WriteFile(MetadataFilePath, data, 0644)
}

func ReadMetadataFromFile() (ContainerMetadata, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	// Read the JSON data from the file
	data, err := os.ReadFile(MetadataFilePath)
	if err != nil {
		// If the file does not exist, return an empty struct instead of an error
		if os.IsNotExist(err) {
			return ContainerMetadata{}, nil
		}
		return ContainerMetadata{}, err
	}

	// Unmarshal the JSON data into the metadata struct
	var metadata ContainerMetadata
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return ContainerMetadata{}, err
	}

	return metadata, nil
}

func GetMetadataHandler(c echo.Context) error {
	// Read the metadata from the file
	metadata, err := ReadMetadataFromFile()
	if err != nil {
		return fmt.Errorf("failed to read metadata from file: %w", err)
	}

	// Return the metadata as JSON
	return c.JSON(http.StatusOK, metadata)
}
