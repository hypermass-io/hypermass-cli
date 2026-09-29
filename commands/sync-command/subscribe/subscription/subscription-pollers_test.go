package subscription

import (
	"context"
	"testing"
)

func testSubscription(streamId string) *Subscription {
	ctx, cancel := context.WithCancel(context.Background())
	return &Subscription{StreamId: streamId, Ctx: ctx, Cancel: cancel}
}

func TestReplaceStoresWhenOldIsCurrent(t *testing.T) {
	pollers := NewSubscriptionPollers()
	old := testSubscription("_abc")
	pollers.Store("_abc", old)
	replacement := testSubscription("_abc")

	if !pollers.Replace("_abc", old, replacement) {
		t.Fatal("expected the replacement to be stored")
	}

	current, _ := pollers.Load("_abc")
	if current != replacement {
		t.Error("expected the replacement to be held for the key")
	}
	if replacement.Ctx.Err() != nil {
		t.Error("expected the stored replacement to be left running")
	}
}

func TestReplaceRefusesWhenOldIsSuperseded(t *testing.T) {
	pollers := NewSubscriptionPollers()
	old := testSubscription("_abc")
	newer := testSubscription("_abc")
	pollers.Store("_abc", newer)
	stale := testSubscription("_abc")

	if pollers.Replace("_abc", old, stale) {
		t.Fatal("expected the stale replacement to be refused")
	}

	current, _ := pollers.Load("_abc")
	if current != newer {
		t.Error("expected the newer subscription to be kept")
	}
	if stale.Ctx.Err() == nil {
		t.Error("expected the refused replacement to be cancelled")
	}
}

func TestReplaceRefusesWhenKeyIsGone(t *testing.T) {
	pollers := NewSubscriptionPollers()
	old := testSubscription("_abc")
	stale := testSubscription("_abc")

	if pollers.Replace("_abc", old, stale) {
		t.Fatal("expected the replacement for a missing key to be refused")
	}

	if _, ok := pollers.Load("_abc"); ok {
		t.Error("expected the key to stay absent")
	}
	if stale.Ctx.Err() == nil {
		t.Error("expected the refused replacement to be cancelled")
	}
}
