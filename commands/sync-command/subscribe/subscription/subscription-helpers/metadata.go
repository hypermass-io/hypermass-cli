package subscription_helpers

import (
	"errors"
	"hypermass-cli/config"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CurrentStateVersion allows for future upcasters
const CurrentStateVersion = 1

// StreamState is the state file of a target directory, recording the stream and direction it belongs to.
type StreamState struct {
	Version       int    `yaml:"version"`
	StreamId      string `yaml:"stream_id"`
	Direction     string `yaml:"direction"`
	LastPayloadId string `yaml:"last_payload_id"`
}

func ReadLastPayloadId(basePath string) string {
	stateFilePath := filepath.Join(basePath, ".hypermass", "state.yaml")

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		// If file doesn't exist, return the magic string "latest" to grab only the latest file
		return "latest"
	}

	var state StreamState
	err = yaml.Unmarshal(data, &state)
	if err != nil {
		log.Printf("⚠️ Warning: Could not parse state file at %s: %v", stateFilePath, err)
		return ""
	}

	return state.LastPayloadId
}

// WriteLastPayloadId records the last payload received by a subscription, along with the stream it belongs to.
func WriteLastPayloadId(basePath string, streamId string, lastPayloadId string) error {
	return writeState(basePath, StreamState{
		Version:       CurrentStateVersion,
		StreamId:      streamId,
		Direction:     config.DirectionSubscription,
		LastPayloadId: lastPayloadId,
	})
}

// readState reads the state file, returning an empty state when there is none.
func readState(basePath string) (StreamState, error) {
	var state StreamState

	data, err := os.ReadFile(filepath.Join(basePath, ".hypermass", "state.yaml"))
	if errors.Is(err, fs.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}

	err = yaml.Unmarshal(data, &state)
	return state, err
}

func writeState(basePath string, state StreamState) error {
	stateDir := filepath.Join(basePath, ".hypermass")
	stateFilePath := filepath.Join(stateDir, "state.yaml")

	// Ensure the directory exists (in case it was deleted)
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		_ = os.MkdirAll(stateDir, 0755)
	}

	stateYaml, err := yaml.Marshal(&state)
	if err != nil {
		return err
	}

	err = os.WriteFile(stateFilePath, stateYaml, 0644)
	if err != nil {
		log.Printf("❌ Unable to write state file: %v", err)
		return err
	}

	return nil
}
