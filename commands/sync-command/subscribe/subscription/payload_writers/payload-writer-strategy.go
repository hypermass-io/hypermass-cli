package payload_writers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

// PayloadWriterStrategy defines the contract for all payload writing methods
type PayloadWriterStrategy interface {
	WritePayload(resp *http.Response, msg messages.PayloadNotificationMessage, folderPath string) error
	// IsPresent reports whether the payload's file or folder is already in the folder. A present payload is left as
	// it is, whatever has happened to it since - deleting it is what marks it consumed.
	IsPresent(msg messages.PayloadNotificationMessage, folderPath string) (bool, error)
}

// GetPayloadWriter returns the appropriate PayloadWriterStrategy based on configuration.
func GetPayloadWriter(strategyType string, streamId string) PayloadWriterStrategy {
	switch strategyType {
	case "folder-with-metadata":
		return &FolderWithMetadataStrategy{}
	case "", "file-per-payload":
		return &FilePerPayloadStrategy{}
	default:
		fmt.Printf("Unknown writer-type '%s' in config stream %s, using default 'file-per-payload' type\n", strategyType, streamId)
		return &FilePerPayloadStrategy{}
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// temporaryPath is where a payload is written before its move into place: its final name behind a random string, so
// no two writes ever share it and a leftover still shows what it was for.
func temporaryPath(folderPath string, finalName string) string {
	return filepath.Join(folderPath, ".hypermass", "temporary", rand.Text()+"-"+finalName)
}
