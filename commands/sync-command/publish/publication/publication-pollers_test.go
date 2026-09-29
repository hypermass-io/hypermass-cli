package publication

import (
	"context"
	"testing"
)

func TestLoadReturnsTheStoredPoller(t *testing.T) {
	pollers := NewPublicationPollers()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	poller := &PublicationPoller{StreamId: "_abc", Ctx: ctx, Cancel: cancel}
	pollers.Store("_abc", poller)

	loaded, ok := pollers.Load("_abc")

	if !ok || loaded != poller {
		t.Error("expected the stored poller to be returned")
	}
}

func TestLoadOfAMissingKeyReportsNotFound(t *testing.T) {
	pollers := NewPublicationPollers()

	loaded, ok := pollers.Load("_abc")

	if ok || loaded != nil {
		t.Error("expected a missing key to report not found")
	}
}
