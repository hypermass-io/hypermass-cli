package sync_command

import (
	"context"
	"hypermass-cli/config"
	"testing"
	"time"
)

func TestAnAnonymousSyncDoesNotWatchCredentials(t *testing.T) {
	profile := config.HypermassProfile{Configuration: config.HypermassConfig{
		SubscriptionConfigurations: []config.SubscriptionConfiguration{{Key: "_abc"}},
	}}
	done := make(chan struct{})

	go func() {
		watchAccountHealth(context.Background(), profile)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected the health watch to return straight away without a key")
	}
}
