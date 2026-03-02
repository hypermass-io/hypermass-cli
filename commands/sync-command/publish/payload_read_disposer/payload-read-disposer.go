package payload_read_disposer

import "fmt"

// PayloadReadDisposerStrategy defines the contract for all file read handlers
// which control how files are disposed of once read
type PayloadReadDisposerStrategy interface {
	DisposeOfPayloadFile(fileToDeletePath string) error
}

// GetPayloadReadDisposer returns the appropriate GetPayloadReadDisposer based on configuration.
func GetPayloadReadDisposer(disposerStrategyType string, streamId string, targetDirectory string) PayloadReadDisposerStrategy {
	switch disposerStrategyType {
	case "", "delete-on-success":
		return &DeleteStrategy{}
	case "move-on-success":
		return &MoveStrategy{
			workingDir: targetDirectory,
		}
	default:
		fmt.Printf("Unknown disposer-type '%s' in config stream %s, using default 'delete-on-success' type\n", disposerStrategyType, streamId)
		return &DeleteStrategy{}
	}
}
