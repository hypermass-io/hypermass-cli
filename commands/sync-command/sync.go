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

	subscriptionPollers := subscription.LoadSubscriptionsFromSettings(ctx, hypermassProfile, commandBus)
	publicationPollers := publish.LoadPublicationPollersFromSettings(ctx, hypermassProfile, commandBus)

	//start the account health process that occasionally polls the API if there's no other activity
	go watchAccountHealth(ctx, hypermassProfile)

	register(commandBus, subscriptionPollers, publicationPollers)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select { // wait until we get a shutdown signal
	case <-interrupt:
		log.Println("OS Interrupt received. Exiting.")
		cancel()
	}

	// waited on after shutdown starts, so streams added while running are included
	subscriptionPollers.WG.Wait()
	publicationPollers.WG.Wait()
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

		return synclock.CommandResponse{Success: true, Data: reportStatus}
	})
}
