package sync_command

import (
	"context"
	"hypermass-cli/commands/sync-command/publish"
	"hypermass-cli/commands/sync-command/subscribe/subscription"
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
	wg.Go(func() { subscription.LoadSubscriptionsFromSettings(ctx, hypermassProfile, commandBus) })
	wg.Go(func() { publish.LoadPublicationPollersFromSettings(ctx, hypermassProfile, commandBus) })

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-interrupt:
		log.Println("OS Interrupt received. Sending cancellation signal to all workers.")
		cancel()
	}

	wg.Wait()
}
