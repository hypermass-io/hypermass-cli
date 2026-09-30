package sync_command

import (
	"context"
	"encoding/json"
	"fmt"
	"hypermass-cli/app_common"
	"hypermass-cli/commands/sync-command/publish"
	"hypermass-cli/commands/sync-command/publish/publication"
	"hypermass-cli/commands/sync-command/subscribe/subscription"
	"hypermass-cli/config"
	"hypermass-cli/config/synclock"
	"log"
	"sync"
)

// reloader applies the configuration file to the running sync, one reload at a time.
type reloader struct {
	mu sync.Mutex

	ctx                 context.Context
	hypermassProfile    config.HypermassProfile
	subscriptionPollers *subscription.SubscriptionPollers
	publicationPollers  *publication.PublicationPollers
}

func registerReload(bus *synclock.CommandBus, r *reloader) {
	bus.Register("reload", r.handle)
}

// handle compares the configuration with the running streams and applies the difference. With noWait the reply
// lists the planned changes and they are applied afterwards.
func (r *reloader) handle(req synclock.CommandRequest) synclock.CommandResponse {
	noWait := req.Params["noWait"] == "true"

	if r.ctx.Err() != nil {
		return synclock.CommandResponse{Success: false, Message: "The sync is shutting down"}
	}

	if noWait {
		if !r.mu.TryLock() {
			return synclock.CommandResponse{Success: false, Message: "A reload is already in progress"}
		}
	} else {
		r.mu.Lock()
	}

	configuration, err := config.ReadConfiguration()
	if err != nil {
		r.mu.Unlock()
		return synclock.CommandResponse{Success: false, Message: err.Error()}
	}

	configuration, warnings := configuration.Deduplicated()
	for _, warning := range warnings {
		log.Println("⚠️ " + warning)
	}

	subscriptionChanges := config.CompareStreams(r.subscriptionPollers.RunningConfigurations(), configuration.SubscriptionConfigurations)
	publicationChanges := config.CompareStreams(r.publicationPollers.RunningConfigurations(), configuration.PublicationConfigurations)

	report := app_common.ReloadReport{
		Streams:   append(plannedStreams(subscriptionChanges, "subscription"), plannedStreams(publicationChanges, "publication")...),
		Unchanged: len(subscriptionChanges.Unchanged) + len(publicationChanges.Unchanged),
		Warnings:  warnings,
	}

	apply := func() {
		defer r.mu.Unlock()

		subscriptionFailures := subscription.ApplyChanges(r.ctx, r.subscriptionPollers, subscriptionChanges, r.hypermassProfile)
		publicationFailures := publish.ApplyChanges(r.ctx, r.publicationPollers, publicationChanges, r.hypermassProfile)

		for i, stream := range report.Streams {
			failures := subscriptionFailures
			if stream.Direction == "publication" {
				failures = publicationFailures
			}

			if err := failures[stream.StreamId]; err != nil {
				report.Streams[i].Error = err.Error()
				log.Printf("Reload: %s %s failed to start: %v", stream.Direction, stream.StreamId, err)
			} else {
				log.Printf("Reload: %s %s %s", stream.Direction, stream.StreamId, stream.Change)
			}
		}
	}

	if !noWait {
		apply()
		report.Completed = true
	}

	//encoded before a no-wait apply starts, as the apply records failures in the report
	reportJson, err := json.Marshal(report)

	if noWait {
		go apply()
	}

	if err != nil {
		return synclock.CommandResponse{Success: false, Message: fmt.Sprintf("Unable to report the reload: %s", err)}
	}

	return synclock.CommandResponse{Success: true, Data: string(reportJson)}
}

func plannedStreams[T config.StreamConfiguration](changes config.StreamChanges[T], direction string) []app_common.ReloadedStream {
	var streams []app_common.ReloadedStream

	for _, entry := range changes.Added {
		streams = append(streams, app_common.ReloadedStream{StreamId: entry.StreamKey(), Direction: direction, Change: "added"})
	}
	for _, entry := range changes.Changed {
		streams = append(streams, app_common.ReloadedStream{StreamId: entry.StreamKey(), Direction: direction, Change: "changed"})
	}
	for _, key := range changes.Removed {
		streams = append(streams, app_common.ReloadedStream{StreamId: key, Direction: direction, Change: "removed"})
	}

	return streams
}
