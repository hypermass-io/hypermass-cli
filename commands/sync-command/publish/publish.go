package publish

import (
	"context"
	"errors"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/helpers"
	"hypermass-cli/commands/sync-command/publish/publication"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"log"
)

// LoadPublicationPollersFromSettings loads and starts running the pollers from settings
func LoadPublicationPollersFromSettings(ctx context.Context, hypermassProfile config.HypermassProfile, commandBus *synclock.CommandBus) *publication.PublicationPollers {

	publicationPollers := publication.NewPublicationPollers()

	for _, publicationConfig := range hypermassProfile.Configuration.PublicationConfigurations {
		err := startPoller(ctx, publicationPollers, publicationConfig, hypermassProfile)
		if err != nil {
			//refused credentials stop the command, because no change to the configuration will fix them and
			//whoever started the sync is here to read the message
			var credentialsRejected *app_errors.CredentialsRejectedError
			if errors.As(err, &credentialsRejected) {
				helpers.StopWithAuthenticationFailure()
			}

			log.Printf("Unable to start publication %s: %v", publicationConfig.Key, err)
		}
	}

	return publicationPollers
}

// errPublishingNeedsAnAccount is why a publication is held as failed without an access key. Short, as status shows it
// in a narrow column.
var errPublishingNeedsAnAccount = errors.New("needs an account (hypermass login)")

// startPoller starts a publication poller and stores it, returning an error if there are issues
func startPoller(parentCtx context.Context, publicationPollers *publication.PublicationPollers, publicationConfig config.PublicationConfiguration, hypermassProfile config.HypermassProfile) error {
	//publishing is for accounts, so without a key the publication is held as failed and the sync carries on
	if !hypermassProfile.Auth.HasKey() {
		publicationPollers.Store(publicationConfig.Key, publication.NewFailedPublicationPoller(parentCtx, publicationConfig, hypermassProfile, errPublishingNeedsAnAccount))
		return errPublishingNeedsAnAccount
	}

	publicationPoller, err := publication.NewPublicationPoller(parentCtx, publicationConfig, hypermassProfile)

	if err != nil {
		if publicationPoller != nil {
			publicationPoller.Cancel()
		}

		publicationPollers.Store(publicationConfig.Key, publication.NewFailedPublicationPoller(parentCtx, publicationConfig, hypermassProfile, err))
		return err
	}

	publicationPollers.Store(publicationConfig.Key, publicationPoller)
	return nil
}

// ApplyChanges stops the removed and changed publications, and any failed entry for an added one, waiting for
// their processors. It then starts the added and changed publications, returning the start error of each that
// failed, by key.
func ApplyChanges(parentCtx context.Context, publicationPollers *publication.PublicationPollers, changes config.StreamChanges[config.PublicationConfiguration], hypermassProfile config.HypermassProfile) map[string]error {
	var stopping []*publication.PublicationPoller
	stop := func(key string) {
		if poller, ok := publicationPollers.Take(key); ok {
			poller.Cancel()
			stopping = append(stopping, poller)
		}
	}

	for _, key := range changes.Removed {
		stop(key)
	}
	for _, entry := range changes.Changed {
		stop(entry.Key)
	}
	for _, entry := range changes.Added {
		stop(entry.Key)
	}

	for _, poller := range stopping {
		poller.ProcessorsWG.Wait()
	}

	failures := make(map[string]error)
	start := func(entry config.PublicationConfiguration) {
		if err := startPoller(parentCtx, publicationPollers, entry, hypermassProfile); err != nil {
			failures[entry.Key] = err
		}
	}

	for _, entry := range changes.Added {
		start(entry)
	}
	for _, entry := range changes.Changed {
		start(entry)
	}

	return failures
}
