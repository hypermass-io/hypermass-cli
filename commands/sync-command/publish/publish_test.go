package publish

import (
	"context"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"testing"
)

// Publishing is for accounts: a keyless sync holds its publications as failed entries and carries on.
func TestAKeylessSyncHoldsPublicationsAsFailed(t *testing.T) {
	profile := config.HypermassProfile{Configuration: config.HypermassConfig{
		PublicationConfigurations: []config.PublicationConfiguration{{Key: "_abc", TargetDirectory: t.TempDir()}},
	}}

	pollers := LoadPublicationPollersFromSettings(context.Background(), profile, synclock.NewCommandBus())

	held, ok := pollers.Load("_abc")
	if !ok {
		t.Fatal("expected the publication to be held as a failed entry")
	}
	if held.StartError != errPublishingNeedsAnAccount || held.ReportingState.Status != "Failed-To-Start" {
		t.Errorf("expected a failed entry needing an account, got %v / %+v", held.StartError, held.ReportingState)
	}
	if held.Ctx.Err() == nil {
		t.Error("expected the held publication to be stopped")
	}
}

func TestAKeylessReloadHoldsAnAddedPublicationAsFailed(t *testing.T) {
	pollers := LoadPublicationPollersFromSettings(context.Background(), config.HypermassProfile{}, synclock.NewCommandBus())
	added := config.PublicationConfiguration{Key: "_abc", TargetDirectory: t.TempDir()}

	failures := ApplyChanges(context.Background(), pollers, config.StreamChanges[config.PublicationConfiguration]{Added: []config.PublicationConfiguration{added}}, config.HypermassProfile{})

	if failures["_abc"] != errPublishingNeedsAnAccount {
		t.Errorf("expected the reload to report the publication needs an account, got %v", failures["_abc"])
	}
}
