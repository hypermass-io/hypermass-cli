package subscription_helpers

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ReadLastPayloadId(filePath string) string {
	lastPayloadFilePath := filepath.Join(filePath, ".hypermass", "last_payload")

	data, err := os.ReadFile(lastPayloadFilePath)

	if err != nil {
		log.Println("Unable to resolve lastPayloadFile: ", lastPayloadFilePath)
	}

	return strings.TrimSpace(string(data))
}

func WriteLastPayloadId(filePath string, lastPayloadId string) error {
	lastPayloadFilePath := filepath.Join(filePath, ".hypermass", "last_payload")

	data := []byte(lastPayloadId)
	err := os.WriteFile(lastPayloadFilePath, data, 0644)

	if err != nil {
		log.Println("Unable to resolve lastPayloadFile: ", lastPayloadFilePath)
		return err
	}

	return nil
}
