package config

import (
	"testing"
)

var newEntry = SubscriptionConfiguration{
	Key:             "_new",
	TargetDirectory: "/home/andy/hypermass/subscriptions/_new",
	WriterType:      "file-per-payload",
}

func assertAppended(t *testing.T, original string, expected string) {
	t.Helper()

	updated, err := appendSubscription([]byte(original), newEntry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(updated) != expected {
		t.Errorf("unexpected result\n--- got ---\n%s\n--- expected ---\n%s", updated, expected)
	}
}

func TestAppendKeepsCommentsAndOrder(t *testing.T) {
	original := `# my hypermass config
base-directory: /home/andy/hypermass
subscription-targets:
  - key: "_first" # the key of the stream
    target-directory: /data/first # override
    writer-type: "file-per-payload"
  # weather feed
  - key: _second
    target-directory: /data/second
publication-sources:
  - key: "_first"
    target-directory: /data/pub
`
	expected := `# my hypermass config
base-directory: /home/andy/hypermass
subscription-targets:
  - key: "_first" # the key of the stream
    target-directory: /data/first # override
    writer-type: "file-per-payload"
  # weather feed
  - key: _second
    target-directory: /data/second
  - key: _new
    target-directory: /home/andy/hypermass/subscriptions/_new
    writer-type: file-per-payload
publication-sources:
  - key: "_first"
    target-directory: /data/pub
`
	assertAppended(t, original, expected)
}

func TestAppendAddsTheListWhenMissing(t *testing.T) {
	original := `base-directory: /home/andy/hypermass
`
	expected := `base-directory: /home/andy/hypermass
subscription-targets:
  - key: _new
    target-directory: /home/andy/hypermass/subscriptions/_new
    writer-type: file-per-payload
`
	assertAppended(t, original, expected)
}

func TestAppendFillsAnEmptyList(t *testing.T) {
	original := `subscription-targets: []
publication-sources: []
`
	expected := `subscription-targets:
  - key: _new
    target-directory: /home/andy/hypermass/subscriptions/_new
    writer-type: file-per-payload
publication-sources: []
`
	assertAppended(t, original, expected)
}

func TestAppendFillsANullList(t *testing.T) {
	original := `subscription-targets:
publication-sources: []
`
	expected := `subscription-targets:
  - key: _new
    target-directory: /home/andy/hypermass/subscriptions/_new
    writer-type: file-per-payload
publication-sources: []
`
	assertAppended(t, original, expected)
}

func TestAppendToAnEmptyFile(t *testing.T) {
	expected := `subscription-targets:
  - key: _new
    target-directory: /home/andy/hypermass/subscriptions/_new
    writer-type: file-per-payload
`
	assertAppended(t, "", expected)
}

func TestAppendQuotesValuesThatNeedIt(t *testing.T) {
	entry := SubscriptionConfiguration{Key: "_new", TargetDirectory: "/data/my: #folder", WriterType: "file-per-payload"}

	updated, err := appendSubscription([]byte("subscription-targets: []\n"), entry)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "subscription-targets:\n  - key: _new\n    target-directory: '/data/my: #folder'\n    writer-type: file-per-payload\n"
	if string(updated) != expected {
		t.Errorf("unexpected result\n%s", updated)
	}
}

func TestAppendRefusesInvalidYaml(t *testing.T) {
	_, err := appendSubscription([]byte("subscription-targets: [\n"), newEntry)

	if err == nil {
		t.Error("expected invalid YAML to be refused")
	}
}

func TestAppendRefusesSubscriptionTargetsThatAreNotAList(t *testing.T) {
	_, err := appendSubscription([]byte("subscription-targets: nonsense\n"), newEntry)

	if err == nil {
		t.Error("expected a non-list subscription-targets to be refused")
	}
}

func TestAppendPublicationAddsToPublicationSources(t *testing.T) {
	original := `subscription-targets:
  - key: _first
    target-directory: /data/first
publication-sources: []
`
	entry := PublicationConfiguration{Key: "_new", TargetDirectory: "/data/hypermass/publications/_new", DisposerType: "delete-on-success"}

	updated, err := appendPublication([]byte(original), entry)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `subscription-targets:
  - key: _first
    target-directory: /data/first
publication-sources:
  - key: _new
    target-directory: /data/hypermass/publications/_new
    disposer-type: delete-on-success
`
	if string(updated) != expected {
		t.Errorf("unexpected result\n--- got ---\n%s\n--- expected ---\n%s", updated, expected)
	}
}
