package subscription_helpers

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CurrentStateVersion allows for future upcasters
const CurrentStateVersion = 1

type SubscriptionState struct {
	Version       int    `yaml:"version"`
	LastPayloadId string `yaml:"last_payload_id"`
}

func ReadLastPayloadId(basePath string) string {
	stateFilePath := filepath.Join(basePath, ".hypermass", "state.yaml")

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		// If file doesn't exist, return the magic string "latest" to grab only the latest file
		return "latest"
	}

	var state SubscriptionState
	err = yaml.Unmarshal(data, &state)
	if err != nil {
		log.Printf("⚠️ Warning: Could not parse state file at %s: %v", stateFilePath, err)
		return ""
	}

	return state.LastPayloadId
}

func WriteLastPayloadId(basePath string, lastPayloadId string) error {
	stateDir := filepath.Join(basePath, ".hypermass")
	stateFilePath := filepath.Join(stateDir, "state.yaml")

	// Ensure the directory exists (in case it was deleted)
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		_ = os.MkdirAll(stateDir, 0755)
	}

	state := SubscriptionState{
		Version:       CurrentStateVersion,
		LastPayloadId: lastPayloadId,
	}

	data, err := yaml.Marshal(&state)
	if err != nil {
		return err
	}

	err = os.WriteFile(stateFilePath, data, 0644)
	if err != nil {
		log.Printf("❌ Unable to write state file: %v", err)
		return err
	}

	return nil
}
