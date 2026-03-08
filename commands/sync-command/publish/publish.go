package publish

import (
	"context"
	"hypermass-cli/commands/sync-command/publish/publication"
	"hypermass-cli/config"
	"log"
	"os"
)

// LoadPublicationPollersFromSettings loads and starts running the pollers from settings
func LoadPublicationPollersFromSettings(ctx context.Context, hypermassProfile config.HypermassProfile) {

	publicationPollers := publication.NewPublicationPollers()

	for _, subscriptionConfig := range hypermassProfile.Configuration.PublicationConfigurations {
		startPoller(ctx, publicationPollers, subscriptionConfig, hypermassProfile)
	}

	publicationPollers.WG.Wait()
}

func startPoller(parentCtx context.Context, publicationPollers *publication.PublicationPollers, publicationConfig config.PublicationConfiguration, hypermassProfile config.HypermassProfile) {
	publicationPoller, err := publication.NewPublicationPoller(parentCtx, publicationConfig, hypermassProfile)

	if err != nil {
		log.Println("Unable to initialise stream")
		if publicationPoller != nil {
			publicationPoller.Cancel()
		}
		os.Exit(1)
	}

	publicationPollers.Store(publicationConfig.Key, *publicationPoller)
}
