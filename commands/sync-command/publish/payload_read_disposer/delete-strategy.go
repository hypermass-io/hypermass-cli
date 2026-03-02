package payload_read_disposer

import (
	"errors"
	"fmt"
	"os"
)

type DeleteStrategy struct{}

// DisposeOfPayloadFile this strategy deletes the completed file
func (s *DeleteStrategy) DisposeOfPayloadFile(fileToDeletePath string) error {

	err := os.Remove(fileToDeletePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If the file doesn't exist, assume it's been deleted and handle like a success
			return nil
		}

		return fmt.Errorf("failed to delete file %s: %w", fileToDeletePath, err)
	}

	return nil
}
