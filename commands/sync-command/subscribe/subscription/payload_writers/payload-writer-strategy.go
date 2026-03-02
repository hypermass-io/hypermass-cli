package payload_writers

import (
	"fmt"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"net/http"
)

// PayloadWriterStrategy defines the contract for all payload writing methods
type PayloadWriterStrategy interface {
	WritePayload(resp *http.Response, msg messages.PayloadNotificationMessage, folderPath string) error
}

// GetPayloadWriter returns the appropriate PayloadWriterStrategy based on configuration.
func GetPayloadWriter(strategyType string, streamId string) PayloadWriterStrategy {
	switch strategyType {
	case "folders-with-metadata":
		return &FolderWithMetadataStrategy{}
	case "", "file-per-payload":
		return &FilePerPayloadStrategy{}
	default:
		fmt.Printf("Unknown writer-type '%s' in config stream %s, using default 'file-per-payload' type\n", strategyType, streamId)
		return &FilePerPayloadStrategy{}
	}
}
