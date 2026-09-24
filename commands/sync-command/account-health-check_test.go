package sync_command

import (
	"hypermass-cli/app_common"
	"hypermass-cli/app_constants"
	"hypermass-cli/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

// withService points the CLI at a stand-in service for the duration of one test, and restores the real
// address afterwards. Without this the calls below go to the live API.
func withService(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	original := app_constants.PublicApiUrl
	app_constants.PublicApiUrl = server.URL + "/api"

	t.Cleanup(func() {
		app_constants.PublicApiUrl = original
		server.Close()
	})
}

// withNothingListening points the CLI at an address that refuses connections, which is what a machine
// that has lost its network sees.
func withNothingListening(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.NotFoundHandler())
	address := server.URL
	server.Close()

	original := app_constants.PublicApiUrl
	app_constants.PublicApiUrl = address + "/api"

	t.Cleanup(func() { app_constants.PublicApiUrl = original })
}

func TestTheFirstPublicationIsPreferred(t *testing.T) {
	profile := config.HypermassProfile{Configuration: config.HypermassConfig{
		PublicationConfigurations:  []config.PublicationConfiguration{{Key: "_publication"}},
		SubscriptionConfigurations: []config.SubscriptionConfiguration{{Key: "_subscription"}},
	}}

	streamId, ok := anyConfiguredStream(profile)

	if !ok || streamId != "_publication" {
		t.Errorf("expected the first publication, got %q ok=%v", streamId, ok)
	}
}

func TestASubscriptionIsUsedWhenNothingIsPublished(t *testing.T) {
	profile := config.HypermassProfile{Configuration: config.HypermassConfig{
		SubscriptionConfigurations: []config.SubscriptionConfiguration{{Key: "_subscription"}},
	}}

	streamId, ok := anyConfiguredStream(profile)

	if !ok || streamId != "_subscription" {
		t.Errorf("expected the first subscription, got %q ok=%v", streamId, ok)
	}
}

// With nothing configured there is no stream to ask about, and the watcher returns without starting.
func TestAnEmptyConfigurationHasNothingToAskAbout(t *testing.T) {
	if _, ok := anyConfiguredStream(config.HypermassProfile{}); ok {
		t.Error("expected no stream from an empty configuration")
	}
}

// A machine that cannot reach the service learns nothing about its account, so the existing alert
// stands. Clearing it would hide a real lockout at the moment the sync is least able to confirm one.
func TestLosingTheConnectionLeavesTheAlertStanding(t *testing.T) {
	withNothingListening(t)
	app_common.RecordCredentialsRejected("key rejected or account locked")

	checkAccountHealth(config.HypermassProfile{}, "_any-stream")

	if _, _, refused := app_common.CredentialsRejection(); !refused {
		t.Error("an unreachable service should leave the existing alert alone")
	}

	app_common.RecordSuccessfulContact()
}

func TestRefusedCredentialsRaiseTheAlert(t *testing.T) {
	withService(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
	app_common.RecordSuccessfulContact()

	checkAccountHealth(config.HypermassProfile{}, "_any-stream")

	if _, _, refused := app_common.CredentialsRejection(); !refused {
		t.Error("a 401 should raise the alert")
	}

	app_common.RecordSuccessfulContact()
}

// A stream this key cannot reach says nothing about the credentials themselves, so the alert clears and
// the contact time is refreshed. Without the refresh the check would repeat every interval forever.
func TestAStreamProblemClearsTheAlert(t *testing.T) {
	for name, status := range map[string]int{
		"forbidden": http.StatusForbidden,
		"not found": http.StatusNotFound,
	} {
		t.Run(name, func(t *testing.T) {
			withService(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) })
			app_common.RecordCredentialsRejected("key rejected or account locked")

			checkAccountHealth(config.HypermassProfile{}, "_any-stream")

			if _, _, refused := app_common.CredentialsRejection(); refused {
				t.Errorf("a %d should clear the alert, the credentials were accepted", status)
			}
		})
	}
}
