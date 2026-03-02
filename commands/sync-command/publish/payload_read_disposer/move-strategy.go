package payload_read_disposer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// MoveStrategy moves completed files to a "completed" folder
type MoveStrategy struct {
	workingDir string
}

// DisposeOfPayloadFile this strategy moves the completed file to a sub folder
func (s *MoveStrategy) DisposeOfPayloadFile(fileToDeletePath string) error {
	completeDir := filepath.Join(s.workingDir, "complete")

	err := os.MkdirAll(completeDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", completeDir, err)
	}

	fileName := filepath.Base(fileToDeletePath)
	destinationPath := filepath.Join(completeDir, fileName)

	err = os.Rename(fileToDeletePath, destinationPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If the file doesn't exist, assume it's been deleted and handle like a success
			return nil
		}

		return fmt.Errorf("failed to move file from %s to %s: %w", fileToDeletePath, destinationPath, err)
	}

	return nil
}
