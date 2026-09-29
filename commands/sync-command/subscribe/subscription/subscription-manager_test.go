package subscription

import (
	"context"
	"hypermass-cli/config"
	"testing"
)

func TestStartSubscriptionStoresAFailedEntry(t *testing.T) {
	pollers := NewSubscriptionPollers()
	//an existing directory without hypermass metadata is refused
	subscriptionConfig := config.SubscriptionConfiguration{Key: "_abc", TargetDirectory: t.TempDir()}

	err := startSubscription(context.Background(), pollers, subscriptionConfig, config.HypermassProfile{})

	if err == nil {
		t.Fatal("expected the start to fail")
	}

	failed, ok := pollers.Load("_abc")
	if !ok {
		t.Fatal("expected a failed entry to be stored")
	}
	if failed.StartError == nil || failed.ReportingState.Status != "Failed-To-Start" {
		t.Errorf("expected a failed-to-start entry, got %+v", failed.ReportingState)
	}
	if failed.Ctx.Err() == nil {
		t.Error("expected the failed entry to be stopped")
	}

	pollers.WG.Wait()
}
