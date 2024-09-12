package handler

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
)

var mutex sync.RWMutex

// The file where the  metadata will be stored
const MetadataFilePath = "./metadata.json"

// Helper function to ensure the metadata file exists
func ensureMetadataFileExists() error {
	if _, err := os.Stat(MetadataFilePath); os.IsNotExist(err) {
		log.Info().Msg("Metadata file does not exist. Creating a new one.")
		file, err := os.Create(MetadataFilePath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create metadata file")
			return err
		}
		defer file.Close()

		if _, err := file.Write([]byte("[]")); err != nil {
			log.Error().Err(err).Msg("Failed to write initial content to metadata file")
			return err
		}
	}
	return nil
}

// Helper function to read and unmarshal metadata
func readAndUnmarshalMetadata(v interface{}) error {
	data, err := os.ReadFile(MetadataFilePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read metadata file")
		return err
	}

	if len(data) == 0 {
		log.Warn().Msg("Metadata file is empty")
		return nil
	}

	if err := json.Unmarshal(data, v); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal metadata")
		return err
	}

	return nil
}

// Helper function to marshal and write metadata
func marshalAndWriteMetadata(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal metadata")
		return err
	}

	if err := os.WriteFile(MetadataFilePath, data, 0644); err != nil {
		log.Error().Err(err).Msg("Failed to write metadata to file")
		return err
	}

	return nil
}

func ReadShortMetadataFromFile() ([]ShortMetadata, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	logger := log.With().Str("func", "ReadShortMetadataFromFile").Logger()

	if err := ensureMetadataFileExists(); err != nil {
		return nil, err
	}

	var metadata []ShortMetadata
	if err := readAndUnmarshalMetadata(&metadata); err != nil {
		return nil, err
	}

	logger.Info().Int("count", len(metadata)).Msg("Successfully read metadata from file")
	return metadata, nil
}

func SaveMetadataToFile(newMetadata ContainerMetadata) error {
	mutex.Lock()
	defer mutex.Unlock()

	logger := log.With().Str("func", "SaveMetadataToFile").Logger()

	if err := ensureMetadataFileExists(); err != nil {
		return err
	}

	var metadataList []ContainerMetadata
	if err := readAndUnmarshalMetadata(&metadataList); err != nil {
		return err
	}

	metadataList = append(metadataList, newMetadata)

	if err := marshalAndWriteMetadata(metadataList); err != nil {
		return err
	}

	logger.Info().Int("count", len(metadataList)).Msg("Successfully saved metadata to file")
	return nil
}

type ShortMetadata struct {
	Tag struct {
		Name   string `json:"name"`
		Digest string `json:"digest"`
	} `json:"tag"`
	Labels struct {
		AllLabels map[string]string `json:"all_labels"`
	} `json:"labels"`
}
