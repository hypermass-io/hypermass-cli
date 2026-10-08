package subscription

import (
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// A payload already in the folder is left as it is and not downloaded again, for either writer.
func TestAPresentPayloadIsLeftAsItIs(t *testing.T) {
	for _, writer := range []struct {
		writerType string
		present    func(folder string) string
	}{
		{"file-per-payload", func(folder string) string {
			path := filepath.Join(folder, "present-payload.json")
			mustWrite(t, path, "edited")
			return path
		}},
		{"folder-with-metadata", func(folder string) string {
			path := filepath.Join(folder, "present-payload", "payload.json")
			mustWrite(t, path, "edited")
			return path
		}},
	} {
		t.Run(writer.writerType, func(t *testing.T) {
			var downloads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				downloads.Add(1)
				_, _ = w.Write([]byte(`{"a":1}`))
			}))
			t.Cleanup(server.Close)

			subscription := testSubscription("_abc")
			subscription.FolderPath = t.TempDir()
			subscription.Writer = payload_writers.GetPayloadWriter(writer.writerType, "_abc")
			subscription.FileQueue = make(chan *messages.PayloadNotificationMessage, 2)
			restarted := false
			subscription.RequestRestart = func(string, error) { restarted = true }

			presentPath := writer.present(subscription.FolderPath)

			subscription.FileQueue <- published("present-payload", server.URL+"/present")
			subscription.FileQueue <- published("next-payload", server.URL+"/next")

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
			if downloads.Load() != 1 {
				t.Errorf("expected only the next payload to be downloaded, got %d downloads", downloads.Load())
			}
			if content, _ := os.ReadFile(presentPath); string(content) != "edited" {
				t.Errorf("expected the present payload left as it was, got %q", content)
			}
		})
	}
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func published(payloadId string, downloadUrl string) *messages.PayloadNotificationMessage {
	msg := notification(payloadId, downloadUrl)
	msg.PublishedTimestamp = "2026-10-08T14:02:11.123Z"
	return msg
}
