package app_common

import (
	"sync"
	"time"
)

// The state of the account credentials, shared by every part of the sync process.
//
// Credentials apply to the whole account, so a refusal on any one call is reported here and communicated to the user.
var (
	accountHealthMutex sync.RWMutex
	credentialsMessage string
	credentialsSince   time.Time
	lastContact        time.Time
)

// RecordCredentialsRejected notes that the service refused the credentials.
//
// Repeated calls keep the time of the first refusal, so the status command can report how long the
// problem has lasted.
func RecordCredentialsRejected(message string) {
	accountHealthMutex.Lock()
	defer accountHealthMutex.Unlock()

	if credentialsMessage == "" {
		credentialsSince = time.Now()
	}

	credentialsMessage = message
}

// RecordSuccessfulContact notes a call that reached the service and was accepted.
//
// This clears any refusal, because one accepted call proves the credentials work. The time is kept so
// that watchAccountHealth can skip its own check while the sync is active.
func RecordSuccessfulContact() {
	accountHealthMutex.Lock()
	defer accountHealthMutex.Unlock()

	credentialsMessage = ""
	credentialsSince = time.Time{}
	lastContact = time.Now()
}

// ContactedWithin reports whether any part of the sync has reached the service within the given period.
func ContactedWithin(window time.Duration) bool {
	accountHealthMutex.RLock()
	defer accountHealthMutex.RUnlock()

	return !lastContact.IsZero() && time.Since(lastContact) < window
}

// CredentialsRejection returns the current refusal message and when it started, if there is one.
func CredentialsRejection() (string, time.Time, bool) {
	accountHealthMutex.RLock()
	defer accountHealthMutex.RUnlock()

	return credentialsMessage, credentialsSince, credentialsMessage != ""
}
