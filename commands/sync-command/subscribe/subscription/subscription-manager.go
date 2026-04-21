package subscription

import (
	"context"
	"fmt"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"log"
	"os"
	"time"
)

// LoadSubscriptionsFromSettings Subscribe to the specified streams
func LoadSubscriptionsFromSettings(parentCtx context.Context, hypermassProfile config.HypermassProfile, commandBus *synclock.CommandBus) {

	subscriptions := NewSubscriptionPollers()
	registerCommands(commandBus, subscriptions)

	for _, subscriptionConfig := range hypermassProfile.Configuration.SubscriptionConfigurations {
		registerSubscription(parentCtx, subscriptions, subscriptionConfig, hypermassProfile)
	}

	subscriptions.WG.Wait()
}

func registerSubscription(parentCtx context.Context, subscriptionPollers *SubscriptionPollers, subscriptionConfig config.SubscriptionConfiguration, hypermassProfile config.HypermassProfile) {
	subscription, err := NewSubscription(parentCtx, subscriptionConfig, hypermassProfile.Auth, time.Duration(0))

	if err != nil {
		log.Println("Unable to initialise stream")
		if subscription != nil {
			subscription.Cancel()
		}
		os.Exit(1)
	}

	subscriptionPollers.Store(subscriptionConfig.Key, subscription)
}

func registerCommands(bus *synclock.CommandBus, subscriptions *SubscriptionPollers) {

	// replay command
	bus.Register("replay", func(req synclock.CommandRequest) synclock.CommandResponse {
		streamId := req.Params["streamId"]
		payloadId := req.Params["payloadId"]

		_, success := subscriptions.Load(streamId)
		if !success {
			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Stream with id '%s' not found", streamId)}
		}

		subscriptions.ResetToPayloadId(streamId, payloadId)

		log.Printf("Jumping stream '%s' to payload '%s'", streamId, payloadId)

		return synclock.CommandResponse{Success: true, Message: "Replay triggered"}
	})
}
