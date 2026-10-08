package sync_command

import (
	"errors"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A sync answering on the control port stops a second one starting, and says which process it is.
func TestASecondSyncIsRefused(t *testing.T) {
	running := func() (*http.Client, *synclock.SyncLock, error) {
		return &http.Client{}, &synclock.SyncLock{PID: 4812, StartedAt: "2026-10-08T14:02:11+01:00"}, nil
	}

	err := refuseIfRunning(running)

	if err == nil || !strings.Contains(err.Error(), "PID 4812") {
		t.Errorf("expected the running sync to be named in a refusal, got %v", err)
	}
}

// No sync answering - none started, or one that has gone and left its lockfile - lets the sync start.
func TestASyncStartsWhenNoneAnswers(t *testing.T) {
	unreachable := func() (*http.Client, *synclock.SyncLock, error) {
		return nil, nil, errors.New("sync process is unreachable")
	}

	if err := refuseIfRunning(unreachable); err != nil {
		t.Errorf("expected the sync to start, got %v", err)
	}
}

// Startup clears what a previous run left part written, and nothing else.
func TestStartupClearsTheTemporaryFolders(t *testing.T) {
	stream := t.TempDir()
	leftover := filepath.Join(stream, ".hypermass", "temporary", "KX4M2Q7PZ3B6RW5DT2NC7YH4GA-p_001m4dh7t0aaaaaaaaaaaaaaaa.json")
	payload := filepath.Join(stream, "p_001m4dh7t0bbbbbbbbbbbbbbbb.json")
	state := filepath.Join(stream, ".hypermass", "state.yaml")
	for _, path := range []string{leftover, payload, state} {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	clearTemporaryFolders([]config.SubscriptionConfiguration{{Key: "_abc", TargetDirectory: stream}})

	if _, err := os.Stat(leftover); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected the leftover to be cleared, got %v", err)
	}
	for _, kept := range []string{payload, state} {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("expected %s to be kept: %v", kept, err)
		}
	}
}
