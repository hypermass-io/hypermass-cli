package subscription_helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func InitialiseAndCheckDirectory(baseFilePath string) error {
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
			return nil //normal case - directory already exists
		} else {
			return errors.New("The target directory is not managed by hypermass")
		}
	} else {
		folderPathError := InitialiseHypermassDirectory(baseFilePath)

		if folderPathError == nil {
			return nil //normal case - created a new directory
		} else {
			return errors.New("The target directory is not managed by hypermass")
		}
	}
}

func InitialiseHypermassDirectory(baseFilePath string) error {
	if err := os.MkdirAll(filepath.Join(baseFilePath, ".hypermass"), 0755); err != nil {
		return fmt.Errorf("unable to create stream path: %w", err)
	} else {
		err := WriteLastPayloadId(baseFilePath, "")

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
