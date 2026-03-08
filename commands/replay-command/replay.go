package replay_command

import (
	"fmt"
	"hypermass-cli/config/synclock"
)

func Replay(streamKey string, payloadID string) {
	params := map[string]string{
		"streamId":   streamKey,
		"payloadId":  payloadID,
		"isEarliest": "true",
	}

	result, err := synclock.Dispatch("replay", params)

	if err != nil {
		fmt.Printf("⚠️ Could not contact hypermass sync process - please check that it is running. Error: %v\n", err)
		// TODO This is where we could put fallback logic to edit the state.yaml manually.
		return
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
	} else {
		fmt.Printf("❌ Command rejected: %s\n", result.Message)
	}
}
