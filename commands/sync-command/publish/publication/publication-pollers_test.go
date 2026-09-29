package publication

import (
	"context"
	"testing"
	"time"
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

func TestCancelledPollerProcessorsFinish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	poller := &PublicationPoller{StreamId: "_abc", Ctx: ctx, Cancel: cancel, FolderPath: t.TempDir(), configurationRead: true}
	poller.ProcessorsWG.Go(poller.pollForFiles)

	cancel()

	finished := make(chan struct{})
	go func() {
		poller.ProcessorsWG.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("expected the poller's processors to finish once cancelled")
	}
}
