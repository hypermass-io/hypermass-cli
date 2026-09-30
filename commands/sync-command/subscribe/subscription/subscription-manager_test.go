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

func TestApplyChangesStopsRemovedSubscriptions(t *testing.T) {
	pollers := NewSubscriptionPollers()
	running := testSubscription("_gone")
	pollers.Store("_gone", running)

	failures := ApplyChanges(context.Background(), pollers, config.StreamChanges[config.SubscriptionConfiguration]{Removed: []string{"_gone"}}, config.HypermassProfile{})

	if len(failures) != 0 {
		t.Errorf("expected no failures, got %v", failures)
	}
	if running.Ctx.Err() == nil {
		t.Error("expected the removed subscription to be stopped")
	}
	if _, ok := pollers.Load("_gone"); ok {
		t.Error("expected the removed subscription to be dropped")
	}
}

func TestApplyChangesRetriesAFailedEntry(t *testing.T) {
	pollers := NewSubscriptionPollers()
	//an existing directory without hypermass metadata is refused
	subscriptionConfig := config.SubscriptionConfiguration{Key: "_abc", TargetDirectory: t.TempDir()}
	_ = startSubscription(context.Background(), pollers, subscriptionConfig, config.HypermassProfile{})
	earlierFailure, _ := pollers.Load("_abc")

	failures := ApplyChanges(context.Background(), pollers, config.StreamChanges[config.SubscriptionConfiguration]{Added: []config.SubscriptionConfiguration{subscriptionConfig}}, config.HypermassProfile{})

	if failures["_abc"] == nil {
		t.Error("expected the retry to report its failure")
	}
	retried, ok := pollers.Load("_abc")
	if !ok || retried == earlierFailure || retried.StartError == nil {
		t.Error("expected the earlier failed entry to be replaced by the retry's")
	}
}

func TestRunningConfigurationsLeavesOutFailedEntries(t *testing.T) {
	pollers := NewSubscriptionPollers()
	running := testSubscription("_running")
	running.SubscriptionConfiguration = config.SubscriptionConfiguration{Key: "_running"}
	pollers.Store("_running", running)
	_ = startSubscription(context.Background(), pollers, config.SubscriptionConfiguration{Key: "_failed", TargetDirectory: t.TempDir()}, config.HypermassProfile{})

	configurations := pollers.RunningConfigurations()

	if len(configurations) != 1 || configurations[0].Key != "_running" {
		t.Errorf("expected only the running subscription, got %v", configurations)
	}
}
