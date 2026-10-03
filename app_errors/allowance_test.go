package app_errors

import (
	"strings"
	"testing"
	"time"
)

func TestASpentFreeDailyAllowanceSaysWhenItResetsAndHowToGetMore(t *testing.T) {
	err := AllowanceUsedUp(true, 6*time.Hour+12*time.Minute)

	if !strings.Contains(err.Error(), "free daily allowance used up, it resets in 6h 12m") ||
		!strings.Contains(err.Error(), "hypermass login") {
		t.Errorf("unexpected message %q", err.Error())
	}
	if err.Summary() != "free daily allowance used up" {
		t.Errorf("unexpected summary %q", err.Summary())
	}
	if err.RetryAfter() != 6*time.Hour+12*time.Minute {
		t.Errorf("expected to wait until the reset, got %v", err.RetryAfter())
	}
}

// With an access key the account's own allowance is the one used up.
func TestASpentAccountAllowancePointsAtUsage(t *testing.T) {
	err := AllowanceUsedUp(false, 5*time.Minute)

	if err.Error() != "account allowance used up, see https://hypermass.io/usage" {
		t.Errorf("unexpected message %q", err.Error())
	}
	if err.Summary() != "waiting on allowance" {
		t.Errorf("unexpected summary %q", err.Summary())
	}
}
