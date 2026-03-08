package subscription

import (
	"context"
	"fmt"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"log"
	"os"
)

// LoadSubscriptionsFromSettings Subscribe to the specified streams
func LoadSubscriptionsFromSettings(parentCtx context.Context, hypermassProfile config.HypermassProfile, commandBus *synclock.CommandBus) {

	subscriptions := NewSubscriptionPollers()
	registerCommands(commandBus, subscriptions)

	for _, subscriptionConfig := range hypermassProfile.Configuration.SubscriptionConfigurations {
		retrySubscriptionWithTimeoutHandler(parentCtx, subscriptions, subscriptionConfig, hypermassProfile)
	}

	subscriptions.WG.Wait()
}

func retrySubscriptionWithTimeoutHandler(parentCtx context.Context, subscriptionPollers *SubscriptionPollers, subscriptionConfig config.SubscriptionConfiguration, hypermassProfile config.HypermassProfile) {
	subscription, err := NewSubscription(parentCtx, subscriptionConfig, hypermassProfile)

	if err != nil {
		log.Println("Unable to initialise stream")
		if subscription != nil {
			subscription.Cancel()
		}
		os.Exit(1)
	}

	subscriptionPollers.Store(subscriptionConfig.Key, *subscription)
}

func registerCommands(bus *synclock.CommandBus, subscriptions *SubscriptionPollers) {

	// replay command
	bus.Register("replay", func(req synclock.CommandRequest) synclock.CommandResponse {
		streamId := req.Params["streamId"]
		payloadId := req.Params["payloadId"]
		earliest := req.Params["isEarliest"]

		_, success := subscriptions.Load(streamId)
		if !success {
			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Stream with id '%s' not found", streamId)}
		}

		log.Printf("Jumping stream '%s' to payload '%s' - is earliest: [%s]", streamId, payloadId, earliest)

		return synclock.CommandResponse{Success: true, Message: "Jumped!"}
	})
}
