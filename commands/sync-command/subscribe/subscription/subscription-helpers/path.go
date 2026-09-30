package subscription_helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// InitialiseAndCheckDirectory prepares the target directory of a stream, creating it if needed. An existing
// directory must be managed by hypermass and belong to the same stream and direction.
func InitialiseAndCheckDirectory(baseFilePath string, streamId string, direction string) error {
	basePathExists, folderPathError := CheckFolderPathExists(baseFilePath)

	if folderPathError != nil {
		return folderPathError
	}

	if basePathExists {
		join := filepath.Join(baseFilePath, ".hypermass")
		metadataDirectoryExists, err := CheckFolderPathExists(join)
		if err != nil {
			return fmt.Errorf("unable to check the folder path exists: %w", err)
		}

		if metadataDirectoryExists {
			return claimDirectory(baseFilePath, streamId, direction) //normal case - directory already exists
		} else {
			return errors.New("The target directory is not managed by hypermass")
		}
	} else {
		folderPathError := InitialiseHypermassDirectory(baseFilePath, streamId, direction)

		if folderPathError == nil {
			return nil //normal case - created a new directory
		} else {
			return errors.New("The target directory is not managed by hypermass")
		}
	}
}

// claimDirectory checks a managed directory belongs to the stream and direction, recording them when the state file
// predates them.
func claimDirectory(baseFilePath string, streamId string, direction string) error {
	state, err := readState(baseFilePath)
	if err != nil {
		return fmt.Errorf("unable to read the state of target directory %s: %w", baseFilePath, err)
	}

	if state.StreamId != "" && state.StreamId != streamId {
		return fmt.Errorf("the target directory %s belongs to stream %s", baseFilePath, state.StreamId)
	}

	if state.Direction != "" && state.Direction != direction {
		return fmt.Errorf("the target directory %s is used by a %s", baseFilePath, state.Direction)
	}

	if state.StreamId == "" || state.Direction == "" {
		state.Version = CurrentStateVersion
		state.StreamId = streamId
		state.Direction = direction
		return writeState(baseFilePath, state)
	}

	return nil
}

func InitialiseHypermassDirectory(baseFilePath string, streamId string, direction string) error {
	if err := os.MkdirAll(filepath.Join(baseFilePath, ".hypermass"), 0755); err != nil {
		return fmt.Errorf("unable to create stream path: %w", err)
	} else {
		err := writeState(baseFilePath, StreamState{Version: CurrentStateVersion, StreamId: streamId, Direction: direction})

		if err != nil {
			return fmt.Errorf("failed to create stream metadata file 'last_payload': %w", err)
		}

		return nil //all okay
	}
}

func CheckFolderPathExists(baseFilePath string) (bool, error) {
	stat, err := os.Stat(baseFilePath)
	basePathExists := err == nil

	if stat != nil && !stat.IsDir() {
		return false, errors.New("Path for stream is not a directory:" + baseFilePath)
	}

	return basePathExists, nil
}
