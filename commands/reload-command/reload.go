package reload_command

import (
	"encoding/json"
	"fmt"
	"hypermass-cli/app_common"
	"hypermass-cli/config/synclock"
	"os"
	"strconv"
	"time"
)

// reloadTimeout allows for streams finishing the file they are transferring before they stop
const reloadTimeout = 5 * time.Minute

func Reload(noWait bool) {
	if _, _, err := synclock.DialSync(); err != nil {
		fmt.Println("⚠️ No sync process running, nothing to reload.")
		os.Exit(1)
	}

	if !noWait {
		fmt.Println("Reloading, waiting for any transfers in progress to finish...")
	}

	result, err := synclock.DispatchWithTimeout("reload", map[string]string{"noWait": strconv.FormatBool(noWait)}, reloadTimeout)
	if err != nil {
		fmt.Printf("⚠️ Could not complete the reload, run 'hypermass status' to see the streams. Error: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Printf("❌ Reload rejected: %s\n", result.Message)
		os.Exit(1)
	}

	var report app_common.ReloadReport
	reportJson, _ := result.Data.(string)
	if err := json.Unmarshal([]byte(reportJson), &report); err != nil {
		fmt.Printf("❌ Failed to read reload response: %v\n", err)
		os.Exit(1)
	}

	if !printReport(report) {
		os.Exit(1)
	}
}

// printReport prints the reload report, returning false when a stream failed to start.
func printReport(report app_common.ReloadReport) bool {
	for _, warning := range report.Warnings {
		fmt.Println("⚠️ " + warning)
	}

	if len(report.Streams) == 0 {
		fmt.Printf("✅ Nothing to reload: %d unchanged\n", report.Unchanged)
		return true
	}

	counts := map[string]int{}
	failed := 0
	for _, stream := range report.Streams {
		counts[stream.Change]++
		if stream.Error != "" {
			failed++
		}
	}

	if !report.Completed {
		fmt.Printf("✅ Reload started: %d to add, %d to change, %d to remove, %d unchanged\n",
			counts["added"], counts["changed"], counts["removed"], report.Unchanged)
	} else if failed > 0 {
		fmt.Printf("❌ Reloaded with %d failed: %d added, %d changed, %d removed, %d unchanged\n",
			failed, counts["added"], counts["changed"], counts["removed"], report.Unchanged)
	} else {
		fmt.Printf("✅ Reloaded: %d added, %d changed, %d removed, %d unchanged\n",
			counts["added"], counts["changed"], counts["removed"], report.Unchanged)
	}

	for _, stream := range report.Streams {
		fmt.Println("  " + describe(stream, report.Completed))
	}

	if !report.Completed {
		fmt.Println("Run 'hypermass status' to see them start.")
	}

	return failed == 0
}

func describe(stream app_common.ReloadedStream, completed bool) string {
	symbols := map[string]string{"added": "+", "changed": "~", "removed": "-"}

	if stream.Error != "" {
		return fmt.Sprintf("✗ %s  %s failed: %s", stream.StreamId, stream.Direction, stream.Error)
	}

	if !completed {
		return fmt.Sprintf("%s %s  %s", symbols[stream.Change], stream.StreamId, stream.Direction)
	}

	outcomes := map[string]string{"added": "started", "changed": "restarted", "removed": "stopped"}
	return fmt.Sprintf("%s %s  %s %s", symbols[stream.Change], stream.StreamId, stream.Direction, outcomes[stream.Change])
}
