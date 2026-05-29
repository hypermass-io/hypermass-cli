package status_command

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/commands/status-command/formatters"
	"hypermass-cli/config/synclock"
	"os"
)

type statusOutput struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Status(format string) {
	result, err := synclock.Dispatch("status", nil)

	if format == "json" {
		if err != nil {
			printJSON(statusOutput{
				Success: false,
				Message: "Could not contact hypermass sync process - please check that it is running.",
				Error:   err.Error(),
			})
			return
		}

		printJSON(statusOutput{
			Success: result.Success,
			Message: result.Message,
			Data:    result.Data,
		})
		return
	}

	if err != nil {
		fmt.Printf("⚠️ Could not contact hypermass sync process - please check that it is running. Error: %v\n", err)
		return
	}

	if result.Success {
		formatters.FormatHumanReadableMessage(result)
	} else {
		fmt.Printf("❌ Command rejected: %s\n", result.Message)
	}
}

func printJSON(output statusOutput) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode status output as JSON: %v\n", err)
	}
}
