package subscription

import (
	"context"
	"fmt"
	subscription_status "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-status"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"log"
	"time"
)

// LoadSubscriptionsFromSettings Subscribe to the specified streams
func LoadSubscriptionsFromSettings(parentCtx context.Context, hypermassProfile config.HypermassProfile, commandBus *synclock.CommandBus) *SubscriptionPollers {

	subscriptions := NewSubscriptionPollers()
	registerCommands(commandBus, subscriptions)

	for _, subscriptionConfig := range hypermassProfile.Configuration.SubscriptionConfigurations {
		err := startSubscription(parentCtx, subscriptions, subscriptionConfig, hypermassProfile)
		if err != nil {
			log.Printf("Unable to start subscription %s: %v", subscriptionConfig.Key, err)
		}
	}

	return subscriptions
}

// startSubscription starts a subscription and registers it in subscriptionPollers
func startSubscription(parentCtx context.Context, subscriptionPollers *SubscriptionPollers, subscriptionConfig config.SubscriptionConfiguration, hypermassProfile config.HypermassProfile) error {
	subscription, err := NewSubscription(parentCtx, subscriptionConfig, hypermassProfile.Auth,
		time.Duration(0), subscription_status.NewInitialState(time.Duration(0)))

	if err != nil {
		if subscription != nil {
			subscription.Cancel()
		}
		subscriptionPollers.Store(subscriptionConfig.Key, NewFailedSubscription(parentCtx, subscriptionConfig, hypermassProfile.Auth, err))
		return err
	}

	subscriptionPollers.Store(subscriptionConfig.Key, subscription)
	return nil
}

func registerCommands(bus *synclock.CommandBus, subscriptions *SubscriptionPollers) {

	// replay command
	bus.Register("replay", func(req synclock.CommandRequest) synclock.CommandResponse {
		streamId := req.Params["streamId"]
		payloadId := req.Params["payloadId"]

		subscription, success := subscriptions.Load(streamId)
		if !success {
			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Stream with id '%s' not found", streamId)}
		}

		if subscription.StartError != nil {
			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Stream with id '%s' failed to start: %s", streamId, subscription.StartError)}
		}

		_, err := subscriptions.ResetToPayloadId(streamId, payloadId)
		if err != nil {
			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Unable to reset stream with id '%s', error: %s", streamId, err)}
		}

		log.Printf("Jumping stream '%s' to payload '%s'", streamId, payloadId)

		return synclock.CommandResponse{Success: true, Message: "Replay triggered"}
	})
}
