package config

import (
	"reflect"
	"testing"
)

func TestDeduplicatedKeepsTheFirstEntryAndWarns(t *testing.T) {
	configuration := HypermassConfig{
		SubscriptionConfigurations: []SubscriptionConfiguration{
			{Key: "_abc", TargetDirectory: "/first"},
			{Key: "_def", TargetDirectory: "/other"},
			{Key: "_abc", TargetDirectory: "/second"},
		},
	}

	deduplicated, warnings := configuration.Deduplicated()

	expected := []SubscriptionConfiguration{
		{Key: "_abc", TargetDirectory: "/first"},
		{Key: "_def", TargetDirectory: "/other"},
	}
	if !reflect.DeepEqual(deduplicated.SubscriptionConfigurations, expected) {
		t.Errorf("expected the first entry for each key, got %v", deduplicated.SubscriptionConfigurations)
	}
	if !reflect.DeepEqual(warnings, []string{"duplicate subscription entry for _abc ignored, the first entry is used"}) {
		t.Errorf("unexpected warnings %v", warnings)
	}
}

func TestDeduplicatedAllowsAKeyInBothLists(t *testing.T) {
	configuration := HypermassConfig{
		SubscriptionConfigurations: []SubscriptionConfiguration{{Key: "_abc"}},
		PublicationConfigurations:  []PublicationConfiguration{{Key: "_abc"}},
	}

	deduplicated, warnings := configuration.Deduplicated()

	if len(deduplicated.SubscriptionConfigurations) != 1 || len(deduplicated.PublicationConfigurations) != 1 {
		t.Errorf("expected the key to be kept in both lists, got %v", deduplicated)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestCompareStreams(t *testing.T) {
	running := []SubscriptionConfiguration{
		{Key: "_same", TargetDirectory: "/same"},
		{Key: "_moved", TargetDirectory: "/old"},
		{Key: "_gone", TargetDirectory: "/gone"},
	}
	configured := []SubscriptionConfiguration{
		{Key: "_new", TargetDirectory: "/new"},
		{Key: "_moved", TargetDirectory: "/moved"},
		{Key: "_same", TargetDirectory: "/same"},
	}

	changes := CompareStreams(running, configured)

	if !reflect.DeepEqual(changes.Added, []SubscriptionConfiguration{{Key: "_new", TargetDirectory: "/new"}}) {
		t.Errorf("unexpected added %v", changes.Added)
	}
	if !reflect.DeepEqual(changes.Changed, []SubscriptionConfiguration{{Key: "_moved", TargetDirectory: "/moved"}}) {
		t.Errorf("unexpected changed %v", changes.Changed)
	}
	if !reflect.DeepEqual(changes.Removed, []string{"_gone"}) {
		t.Errorf("unexpected removed %v", changes.Removed)
	}
	if !reflect.DeepEqual(changes.Unchanged, []string{"_same"}) {
		t.Errorf("unexpected unchanged %v", changes.Unchanged)
	}
}

func TestCompareStreamsTreatsAnySettingAsAChange(t *testing.T) {
	running := []PublicationConfiguration{{Key: "_abc", TargetDirectory: "/pub", DisposerType: "delete-on-success"}}
	configured := []PublicationConfiguration{{Key: "_abc", TargetDirectory: "/pub", DisposerType: "move-on-success"}}

	changes := CompareStreams(running, configured)

	if len(changes.Changed) != 1 || len(changes.Unchanged) != 0 {
		t.Errorf("expected a disposer change to count as changed, got %+v", changes)
	}
}
