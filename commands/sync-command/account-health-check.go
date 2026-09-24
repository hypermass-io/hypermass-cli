package sync_command

import (
	"context"
	"errors"
	"hypermass-cli/app_common"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/publish/publication/publication-helpers"
	"hypermass-cli/config"
	"time"
)

// How long the sync waits between credential checks.
const accountHealthInterval = 15 * time.Minute

// watchAccountHealth periodically checks that the credentials work, if no other activity has done this.
//
// The publication poller in particular is liable to long periods between connections to the server.
//
// One goroutine serves the whole process. It skips its check whenever something else has contacted the
// service within the interval, so an active sync makes no extra calls.
//
// The stream it asks about is chosen once, at startup. Any stream works, because what matters is what
// the answer says about the credentials.
func watchAccountHealth(ctx context.Context, hypermassProfile config.HypermassProfile) {
	streamId, ok := anyConfiguredStream(hypermassProfile)
	if !ok {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(accountHealthInterval):
			if app_common.ContactedWithin(accountHealthInterval) {
				continue
			}

			checkAccountHealth(hypermassProfile, streamId)
		}
	}
}

// checkAccountHealth makes one call and records what the answer says about the credentials.
//   - success: the credentials work
//   - 401: the credentials were refused, whichever stream was named
//   - 403 or 404: the credentials work and this one stream is unavailable to them
//   - anything else, such as a network failure: no answer was received, state left unchanged
func checkAccountHealth(hypermassProfile config.HypermassProfile, streamId string) {
	_, err := publication_helpers.GetConfigurationForStream(hypermassProfile, streamId)

	if err == nil {
		app_common.RecordSuccessfulContact()
		return
	}

	var credentialsRejected *app_errors.CredentialsRejectedError
	if errors.As(err, &credentialsRejected) {
		app_common.RecordCredentialsRejected(credentialsRejected.Summary())
		return
	}

	var accessDenied *app_errors.StreamAccessDeniedError
	var streamNotFound *app_errors.StreamNotFoundError

	if errors.As(err, &accessDenied) || errors.As(err, &streamNotFound) {
		app_common.RecordSuccessfulContact()
	}
}

// anyConfiguredStream returns the first stream in the configuration.
func anyConfiguredStream(hypermassProfile config.HypermassProfile) (string, bool) {
	publications := hypermassProfile.Configuration.PublicationConfigurations
	if len(publications) > 0 {
		return publications[0].Key, true
	}

	subscriptions := hypermassProfile.Configuration.SubscriptionConfigurations
	if len(subscriptions) > 0 {
		return subscriptions[0].Key, true
	}

	return "", false
}
