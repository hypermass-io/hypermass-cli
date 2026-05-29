package sync_command

import (
	"context"
	"fmt"
	"hypermass-cli/commands/sync-command/publish"
	"hypermass-cli/commands/sync-command/publish/publication"
	"hypermass-cli/commands/sync-command/subscribe/subscription"
	"hypermass-cli/commands/sync-command/sync-status"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func SyncRunner(hypermassProfile config.HypermassProfile) {
	commandBus := synclock.NewCommandBus()
	controlServer, err := synclock.NewControlServer()
	if err != nil {
		log.Fatalf("unable to create the Control server for sync command: %v", err)
	}
	controlServer.Bus = commandBus
	err = controlServer.Start()
	if err != nil {
		log.Fatalf("unable to start the Control server for sync command: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A WaitGroup is used to block the main function until all background goroutines are done.
	var wg sync.WaitGroup
	subscriptionPollers := subscription.LoadSubscriptionsFromSettings(ctx, hypermassProfile, commandBus)
	wg.Go(func() { subscriptionPollers.WG.Wait() })

	publicationPollers := publish.LoadPublicationPollersFromSettings(ctx, hypermassProfile, commandBus)
	wg.Go(func() { publicationPollers.WG.Wait() })

	register(commandBus, subscriptionPollers, publicationPollers)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select { // wait until we get a shutdown signal
	case <-interrupt:
		log.Println("OS Interrupt received. Exiting.")
		cancel()
	}

	wg.Wait() // wait until workers have actually stopped
}

func register(bus *synclock.CommandBus, subscriptionPollers *subscription.SubscriptionPollers, publicationPollers *publication.PublicationPollers) {

	// replay command
	bus.Register("status", func(req synclock.CommandRequest) synclock.CommandResponse {

		subscriptionSnapshot := subscriptionPollers.Snapshot()
		publicationsSnapshot := publicationPollers.Snapshot()

		reportStatus, err := sync_status.ReportStatus(subscriptionSnapshot, publicationsSnapshot)

		if err != nil {
			return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Unable to get status: %s", err)}
		}

		log.Printf("Reporting status")

		return synclock.CommandResponse{Success: true, Data: reportStatus}
	})
}
