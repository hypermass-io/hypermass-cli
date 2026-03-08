package subscription

import (
	"context"
	"hypermass-cli/config"
	"log"
	"os"
)

// LoadSubscriptionsFromSettings Subscribe to the specified streams
func LoadSubscriptionsFromSettings(parentCtx context.Context, hypermassProfile config.HypermassProfile) {

	subscriptions := NewSubscriptionPollers()

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
