package sync_command

import (
	"fmt"
	"hypermass-cli/commands/sync-command/helpers"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// refuseIfRunning stops a second sync from starting while one answers on the lockfile's control port. A lockfile left
// by a sync that has gone answers nothing, so it does not block.
func refuseIfRunning(dial func() (*http.Client, *synclock.SyncLock, error)) error {
	if _, lock, err := dial(); err == nil {
		return fmt.Errorf("a sync is already running (PID %d, started %s) - stop it before starting another",
			lock.PID, lock.StartedAt)
	}
	return nil
}

// clearTemporaryFolders removes anything a previous run left part written. Only safe before any subscription starts,
// with no other sync running.
func clearTemporaryFolders(subscriptions []config.SubscriptionConfiguration) {
	for _, subscription := range subscriptions {
		temporary := filepath.Join(helpers.GetStreamPathFromConfig(subscription.TargetDirectory), ".hypermass", "temporary")
		if err := os.RemoveAll(temporary); err != nil {
			log.Printf("Unable to clear the temporary folder %s: %v", temporary, err)
		}
	}
}
