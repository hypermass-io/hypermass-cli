package subscription

import (
	"context"
	"hypermass-cli/config"
	"log"
	"os"
)

// LoadSubscriptionsFromSettings Subscribe to the specified streams
func LoadSubscriptionsFromSettings(parentCtx context.Context, globalConfig config.HypermassConfig) {

	subscriptions := NewSubscriptionPollers()

	for _, subscriptionConfig := range globalConfig.SubscriptionConfigurations {
		retrySubscriptionWithTimeoutHandler(parentCtx, subscriptions, subscriptionConfig, globalConfig)
	}

	subscriptions.WG.Wait()
}

func retrySubscriptionWithTimeoutHandler(parentCtx context.Context, subscriptionPollers *SubscriptionPollers, subscriptionConfig config.SubscriptionConfiguration, globalConfig config.HypermassConfig) {
	subscription, err := NewSubscription(parentCtx, subscriptionConfig, globalConfig)

	if err != nil {
		log.Println("Unable to initialise stream")
		if subscription != nil {
			subscription.Cancel()
		}
		os.Exit(1)
	}

	subscriptionPollers.Store(subscriptionConfig.Key, *subscription)
}
