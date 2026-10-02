package subscription

import (
	"hypermass-cli/app_constants"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serviceAnswering points the subscription authorisation at a local server answering with the status.
func serviceAnswering(t *testing.T, status int) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte(`{"connectionUrl":"ws://localhost/unused"}`))
		}
	}))
	t.Cleanup(server.Close)

	original := app_constants.BulkAuthenticationApiUrl
	app_constants.BulkAuthenticationApiUrl = server.URL + "/api/data/bulk/authorise/infochannel"
	t.Cleanup(func() { app_constants.BulkAuthenticationApiUrl = original })
}

// A rejoining subscription whose last payload is no longer known resumes from the earliest available.
func TestAnUnknownLastPayloadResumesFromTheEarliest(t *testing.T) {
	subscription := testSubscription("_abc")
	subscription.FolderPath = t.TempDir()
	subscription.LastPayloadId = "a-payload-long-since-removed"

	subscription.resumeFromEarliest()

	if resumesFrom := subscriptionhelpers.ReadLastPayloadId(subscription.FolderPath); resumesFrom != "earliest" {
		t.Errorf("expected to resume from earliest, got %q", resumesFrom)
	}
}

// A replay to an id the stream never had is refused, leaving the subscription as it was.
func TestReplayToAnUnknownPayloadChangesNothing(t *testing.T) {
	serviceAnswering(t, http.StatusUnprocessableEntity)
	bus := synclock.NewCommandBus()
	subscriptions := LoadSubscriptionsFromSettings(t.Context(), config.HypermassProfile{}, bus)
	running := testSubscription("_abc")
	subscriptions.Store("_abc", running)

	response := bus.Dispatch(synclock.CommandRequest{Command: "replay", Params: map[string]string{"streamId": "_abc", "payloadId": "a-mistyped-id"}})

	if response.Success || !strings.Contains(response.Message, "not known") {
		t.Errorf("expected the replay to be refused, got %+v", response)
	}
	if running.Ctx.Err() != nil {
		t.Error("expected the subscription to be left running")
	}
}
