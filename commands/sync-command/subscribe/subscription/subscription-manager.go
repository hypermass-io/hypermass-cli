package subscription

import (
	"context"
	"errors"
	"fmt"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/subscribe"
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

// ApplyChanges stops the removed and changed subscriptions, and any failed entry for an added one, waiting for
// their processors. It then starts the added and changed subscriptions, returning the start error of each that
// failed, by key.
func ApplyChanges(parentCtx context.Context, subscriptionPollers *SubscriptionPollers, changes config.StreamChanges[config.SubscriptionConfiguration], hypermassProfile config.HypermassProfile) map[string]error {
	var stopping []*Subscription
	stop := func(key string) {
		if subscription, ok := subscriptionPollers.Take(key); ok {
			subscription.Cancel()
			stopping = append(stopping, subscription)
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

	for _, subscription := range stopping {
		subscription.ProcessorsWG.Wait()
	}

	failures := make(map[string]error)
	start := func(entry config.SubscriptionConfiguration) {
		if err := startSubscription(parentCtx, subscriptionPollers, entry, hypermassProfile); err != nil {
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

		//checked first, so an id the stream never had changes nothing rather than replaying from the start
		if _, err := subscribe.GetAuthorizedSubscriptionUrl(subscription.Auth, streamId, payloadId); err != nil {
			var unknownAnchor *app_errors.UnknownAnchorError
			if errors.As(err, &unknownAnchor) {
				return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Payload '%s' is not known on stream '%s', nothing changed", payloadId, streamId)}
			}

			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Unable to check payload '%s' (%s), nothing changed", payloadId, err)}
		}

		_, err := subscriptions.ResetToPayloadId(streamId, payloadId)
		if err != nil {
			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Unable to reset stream with id '%s', error: %s", streamId, err)}
		}

		log.Printf("Jumping stream '%s' to payload '%s'", streamId, payloadId)

		return synclock.CommandResponse{Success: true, Message: "Replay triggered"}
	})
}
