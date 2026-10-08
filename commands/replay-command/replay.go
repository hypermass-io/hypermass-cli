package replay_command

import (
	"fmt"
	"hypermass-cli/app_common"
	"hypermass-cli/config/synclock"
	"os"
)

func Replay(streamKey string, payloadID string) {
	params := map[string]string{
		"streamId":   streamKey,
		"payloadId":  payloadID,
		"isEarliest": "true",
	}

	result, err := synclock.Dispatch("replay", params)

	if err != nil {
		fmt.Println("⚠️ Could not contact hypermass sync process - please check that it is running.")
		if app_common.Verbose {
			fmt.Printf("Error: %v\n", err)
		}
		// TODO This is where we could put fallback logic to edit the state.yaml manually.
		//  needs some thought - is this a good idea?
		//  e.g. how to differentiate latest, earliest and not-yet-initialised
		os.Exit(1)
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
	} else {
		fmt.Printf("❌ Command rejected: %s\n", result.Message)
		os.Exit(1)
	}
}
