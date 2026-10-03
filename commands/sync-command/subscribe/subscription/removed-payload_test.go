package subscription

import (
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A payload removed by its publisher while it waited in the queue is skipped, and the subscription carries on.
func TestARemovedPayloadIsSkipped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/removed" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"a":1}`))
	}))
	t.Cleanup(server.Close)

	subscription := testSubscription("_abc")
	subscription.FolderPath = t.TempDir()
	subscription.Writer = payload_writers.GetPayloadWriter("file-per-payload", "_abc")
	subscription.FileQueue = make(chan *messages.PayloadNotificationMessage, 2)
	restarted := false
	subscription.RequestRestart = func(string, error) { restarted = true }

	subscription.FileQueue <- notification("removed-payload", server.URL+"/removed")
	subscription.FileQueue <- notification("next-payload", server.URL+"/next")

	go subscription.StartFileQueueProcessor()
	t.Cleanup(subscription.Cancel)

	deadline := time.Now().Add(5 * time.Second)
	for subscriptionhelpers.ReadLastPayloadId(subscription.FolderPath) != "next-payload" {
		if time.Now().After(deadline) {
			t.Fatalf("expected the next payload to be received, last recorded %q",
				subscriptionhelpers.ReadLastPayloadId(subscription.FolderPath))
		}
		time.Sleep(10 * time.Millisecond)
	}

	if restarted {
		t.Error("expected the subscription to carry on rather than restart")
	}
	if _, err := os.Stat(filepath.Join(subscription.FolderPath, "next-payload.json")); err != nil {
		t.Errorf("expected the next payload to be written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(subscription.FolderPath, "removed-payload.json")); err == nil {
		t.Error("expected nothing written for the removed payload")
	}
}

func notification(payloadId string, downloadUrl string) *messages.PayloadNotificationMessage {
	return &messages.PayloadNotificationMessage{
		Type:          "PayloadNotificationMessage",
		StreamId:      "_abc",
		PayloadId:     payloadId,
		FileExtension: "json",
		DownloadUrl:   downloadUrl,
	}
}
