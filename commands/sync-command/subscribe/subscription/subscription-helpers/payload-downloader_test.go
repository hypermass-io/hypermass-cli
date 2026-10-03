package subscription_helpers

import (
	"errors"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/subscribe/messages"
	"hypermass-cli/commands/sync-command/subscribe/subscription/payload_writers"
	"hypermass-cli/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A download refused for allowance waits as long as the service says, which is until the reset for an address.
func TestASpentAllowanceOnDownloadWaitsUntilTheReset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusPaymentRequired)
	}))
	t.Cleanup(server.Close)

	err := DownloadPayload(config.HypermassAuth{}, t.TempDir(), payload_writers.GetPayloadWriter("file-per-payload", "_abc"),
		messages.PayloadNotificationMessage{PayloadId: "a-payload", StreamId: "_abc", FileExtension: "json", DownloadUrl: server.URL})

	var allowance *app_errors.InsufficientAllowanceError
	if !errors.As(err, &allowance) || !allowance.Anonymous || allowance.RetryAfter() != time.Hour {
		t.Errorf("expected the free daily allowance, resetting in an hour, got %+v", err)
	}
}
