package app_errors

import (
	"testing"
	"time"
)

// Each mismatch of the same payload in a row waits longer, settling at a day, so a payload that never matches costs
// one download a day.
func TestAHashMismatchBacksOffToADay(t *testing.T) {
	for mismatches, expected := range map[int]time.Duration{
		1: time.Minute,
		2: 10 * time.Minute,
		3: time.Hour,
		4: 24 * time.Hour,
		9: 24 * time.Hour,
	} {
		err := &PayloadHashMismatchError{Mismatches: mismatches}
		if err.RetryAfter() != expected {
			t.Errorf("mismatch %d: expected %s, got %s", mismatches, expected, err.RetryAfter())
		}
	}
}
