package app_common

import (
	"testing"
	"time"
)

func TestNothingIsReportedUntilSomethingIsRefused(t *testing.T) {
	RecordSuccessfulContact()

	if _, _, refused := CredentialsRejection(); refused {
		t.Error("a healthy sync should report nothing")
	}
}

func TestARefusalIsReportedUntilSomethingWorks(t *testing.T) {
	RecordSuccessfulContact()
	RecordCredentialsRejected("key rejected or account locked")

	message, _, refused := CredentialsRejection()
	if !refused || message != "key rejected or account locked" {
		t.Errorf("expected the refusal to be reported, got %q refused=%v", message, refused)
	}

	//one accepted call shows the credentials work, whichever component made it
	RecordSuccessfulContact()

	if _, _, refused := CredentialsRejection(); refused {
		t.Error("a successful call should clear the refusal")
	}
}

// The banner reports how long the problem has lasted, so repeated refusals keep the first time.
func TestRepeatedRefusalsKeepTheFirstTime(t *testing.T) {
	RecordSuccessfulContact()
	RecordCredentialsRejected("key rejected or account locked")

	_, first, _ := CredentialsRejection()

	time.Sleep(2 * time.Millisecond)
	RecordCredentialsRejected("key rejected or account locked")

	_, second, _ := CredentialsRejection()

	if !first.Equal(second) {
		t.Errorf("expected the first refusal time to be kept, got %v then %v", first, second)
	}
}

// The periodic check skips its call while anything else is reaching the service, so an active sync
// makes no extra calls.
func TestContactIsRememberedSoTheCheckCanStandDown(t *testing.T) {
	RecordSuccessfulContact()

	if !ContactedWithin(time.Minute) {
		t.Error("a successful call should count as recent contact")
	}

	if ContactedWithin(time.Nanosecond) {
		t.Error("contact should not count as recent outside the window")
	}
}

func TestRefusingAgainAfterRecoveryStartsTheClockOver(t *testing.T) {
	RecordSuccessfulContact()
	RecordCredentialsRejected("key rejected or account locked")
	_, first, _ := CredentialsRejection()

	RecordSuccessfulContact()
	time.Sleep(2 * time.Millisecond)
	RecordCredentialsRejected("key rejected or account locked")

	_, second, _ := CredentialsRejection()

	if !second.After(first) {
		t.Errorf("expected a new refusal to restart the clock, got %v then %v", first, second)
	}
}
