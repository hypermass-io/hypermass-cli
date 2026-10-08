package subscription

import (
	"errors"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	"hypermass-cli/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Mismatches of the same payload are counted in a row, and a different payload starts the count again.
func TestASubscriptionCountsMismatchesInARow(t *testing.T) {
	subscription := testSubscription("_abc")

	for expected := 1; expected <= 3; expected++ {
		mismatch := &app_errors.PayloadHashMismatchError{PayloadId: "p_first"}
		subscription.countHashMismatch(mismatch)
		if mismatch.Mismatches != expected {
			t.Errorf("expected mismatch %d of the same payload, got %d", expected, mismatch.Mismatches)
		}
	}

	other := &app_errors.PayloadHashMismatchError{PayloadId: "p_second"}
	subscription.countHashMismatch(other)
	if other.Mismatches != 1 {
		t.Errorf("expected a different payload to start again, got %d", other.Mismatches)
	}
}

// A mismatch reaches the subscription as itself, not as a general download failure, so it can be counted.
func TestADownloadedMismatchKeepsItsType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	algorithm, wrong := "sha256", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	err := subscriptionhelpers.DownloadPayload(config.HypermassAuth{}, t.TempDir(),
		payload_writers.GetPayloadWriter("file-per-payload", "_abc"),
		messages.PayloadNotificationMessage{PayloadId: "p_001m4dh7t0aaaaaaaaaaaaaaaa", StreamId: "_abc",
			FileExtension: "json", PublishedTimestamp: "2026-10-08T14:02:11.123Z", ContentHashAlgorithm: &algorithm,
			ContentHash: &wrong, DownloadUrl: server.URL})

	var mismatch *app_errors.PayloadHashMismatchError
	if !errors.As(err, &mismatch) || mismatch.PayloadId != "p_001m4dh7t0aaaaaaaaaaaaaaaa" {
		t.Errorf("expected a hash mismatch for the payload, got %v", err)
	}
}

// A mismatch while downloading is counted onto the error the subscription restarts with, so the restart waits longer.
func TestARepeatedMismatchRestartsWithTheCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	subscription := testSubscription("_abc")
	subscription.FolderPath = t.TempDir()
	subscription.Writer = payload_writers.GetPayloadWriter("file-per-payload", "_abc")
	subscription.FileQueue = make(chan *messages.PayloadNotificationMessage, 1)
	subscription.hashMismatch = hashMismatch{payloadId: "p_001m4dh7t0aaaaaaaaaaaaaaaa", count: 2}
	restartedWith := make(chan error, 1)
	subscription.RequestRestart = func(_ string, err error) { restartedWith <- err }

	algorithm, wrong := "sha256", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	msg := published("p_001m4dh7t0aaaaaaaaaaaaaaaa", server.URL)
	msg.ContentHashAlgorithm, msg.ContentHash = &algorithm, &wrong
	subscription.FileQueue <- msg

	go subscription.StartFileQueueProcessor()
	t.Cleanup(subscription.Cancel)

	select {
	case err := <-restartedWith:
		var mismatch *app_errors.PayloadHashMismatchError
		if !errors.As(err, &mismatch) || mismatch.Mismatches != 3 {
			t.Errorf("expected the third mismatch in a row, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("expected the subscription to restart")
	}
}

// Getting past a payload clears the count, so a later mismatch - even of the same payload, replayed - starts from one.
func TestGettingPastAPayloadClearsTheCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	subscription := testSubscription("_abc")
	subscription.FolderPath = t.TempDir()
	subscription.Writer = payload_writers.GetPayloadWriter("file-per-payload", "_abc")
	subscription.FileQueue = make(chan *messages.PayloadNotificationMessage, 1)
	subscription.hashMismatch = hashMismatch{payloadId: "p_001m4dh7t0aaaaaaaaaaaaaaaa", count: 3}

	subscription.FileQueue <- published("p_001m4dh7t0aaaaaaaaaaaaaaaa", server.URL)

	stopped := make(chan struct{})
	go func() {
		subscription.StartFileQueueProcessor()
		close(stopped)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for subscriptionhelpers.ReadLastPayloadId(subscription.FolderPath) != "p_001m4dh7t0aaaaaaaaaaaaaaaa" {
		if time.Now().After(deadline) {
			t.Fatal("expected the payload to be received")
		}
		time.Sleep(10 * time.Millisecond)
	}

	//read once the processor has stopped, so the check cannot race it
	subscription.Cancel()
	<-stopped

	if subscription.hashMismatch != (hashMismatch{}) {
		t.Errorf("expected the count to be cleared, got %+v", subscription.hashMismatch)
	}
}
