package publish

import (
	"context"
	"hypermass-cli/commands/sync-command/publish/publication"
	"hypermass-cli/config"
	"log"
	"os"
)

// LoadPublicationPollersFromSettings loads and starts running the pollers from settings
func LoadPublicationPollersFromSettings(ctx context.Context, globalConfig config.HypermassConfig) {

	publicationPollers := publication.NewPublicationPollers()

	for _, subscriptionConfig := range globalConfig.PublicationConfigurations {
		startPoller(ctx, publicationPollers, subscriptionConfig, globalConfig)
	}

	publicationPollers.WG.Wait()
}

func startPoller(parentCtx context.Context, publicationPollers *publication.PublicationPollers, publicationConfig config.PublicationConfiguration, globalConfig config.HypermassConfig) {
	publicationPoller, err := publication.NewPublicationPoller(parentCtx, publicationConfig, globalConfig)

	if err != nil {
		log.Println("Unable to initialise stream")
		if publicationPoller != nil {
			publicationPoller.Cancel()
		}
		os.Exit(1)
	}

	publicationPollers.Store(publicationConfig.Key, *publicationPoller)
}
