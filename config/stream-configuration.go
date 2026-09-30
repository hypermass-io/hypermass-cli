package config

import "fmt"

// StreamConfiguration is a subscription or publication entry, identified by its stream key.
type StreamConfiguration interface {
	comparable
	StreamKey() string
}

func (c SubscriptionConfiguration) StreamKey() string {
	return c.Key
}

func (c PublicationConfiguration) StreamKey() string {
	return c.Key
}

// Deduplicated keeps the first entry for each stream key in each list, returning a warning for every entry
// ignored. Subscriptions and publications are separate lists, so a key may appear once in each.
func (c HypermassConfig) Deduplicated() (HypermassConfig, []string) {
	subscriptions, subscriptionWarnings := firstEntryPerKey(c.SubscriptionConfigurations, "subscription")
	publications, publicationWarnings := firstEntryPerKey(c.PublicationConfigurations, "publication")

	deduplicated := HypermassConfig{
		SubscriptionConfigurations: subscriptions,
		PublicationConfigurations:  publications,
	}

	return deduplicated, append(subscriptionWarnings, publicationWarnings...)
}

func firstEntryPerKey[T StreamConfiguration](entries []T, direction string) ([]T, []string) {
	seen := make(map[string]bool)
	var kept []T
	var warnings []string

	for _, entry := range entries {
		if seen[entry.StreamKey()] {
			warnings = append(warnings, fmt.Sprintf("duplicate %s entry for %s ignored, the first entry is used", direction, entry.StreamKey()))
			continue
		}

		seen[entry.StreamKey()] = true
		kept = append(kept, entry)
	}

	return kept, warnings
}

// StreamChanges is the difference between the running entries and the configured ones, keyed by stream.
type StreamChanges[T StreamConfiguration] struct {
	Added     []T      // configured, not running
	Changed   []T      // running with different settings, holding the configured entry
	Removed   []string // running, no longer configured
	Unchanged []string // running with the configured settings
}

// CompareStreams compares running entries with configured ones. Both must be free of duplicate keys. Results
// follow the configured order, with removals in the running order.
func CompareStreams[T StreamConfiguration](running []T, configured []T) StreamChanges[T] {
	var changes StreamChanges[T]

	runningByKey := make(map[string]T, len(running))
	for _, entry := range running {
		runningByKey[entry.StreamKey()] = entry
	}

	configuredKeys := make(map[string]bool, len(configured))
	for _, entry := range configured {
		configuredKeys[entry.StreamKey()] = true

		current, isRunning := runningByKey[entry.StreamKey()]
		switch {
		case !isRunning:
			changes.Added = append(changes.Added, entry)
		case current != entry:
			changes.Changed = append(changes.Changed, entry)
		default:
			changes.Unchanged = append(changes.Unchanged, entry.StreamKey())
		}
	}

	for _, entry := range running {
		if !configuredKeys[entry.StreamKey()] {
			changes.Removed = append(changes.Removed, entry.StreamKey())
		}
	}

	return changes
}
