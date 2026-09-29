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
	"os"
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

			log.Println("Unable to initialise stream")
			os.Exit(1)
		}
	}

	return publicationPollers
}

// startPoller starts a publication poller and stores it, returning an error if there are issues
func startPoller(parentCtx context.Context, publicationPollers *publication.PublicationPollers, publicationConfig config.PublicationConfiguration, hypermassProfile config.HypermassProfile) error {
	publicationPoller, err := publication.NewPublicationPoller(parentCtx, publicationConfig, hypermassProfile)

	if err != nil {
		if publicationPoller != nil {
			publicationPoller.Cancel()
		}
		return err
	}

	publicationPollers.Store(publicationConfig.Key, publicationPoller)
	return nil
}
