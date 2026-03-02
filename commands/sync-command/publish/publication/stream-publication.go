package publication

import (
	"context"
	"errors"
	"fmt"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/helpers"
	"hypermass-cli/commands/sync-command/publish/payload_read_disposer"
	publication_helpers "hypermass-cli/commands/sync-command/publish/publication/publication-helpers"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	"hypermass-cli/config"
	"log"
	"time"
)

// MinFileAge is the minimum time that must pass since modification (mtime) before a file is considered write-complete
//
//	and therefore ready for the PublicationPoller to upload. This prevents reading incomplete writes.
const MinFileAge = 5 * time.Second

// MaxFileSize is the maximum allowed size for a file, set to 100 MB initially.
const MaxFileSize = 100 * 1024 * 1024

type PublicationPoller struct {
	StreamId string
	Ctx      context.Context
	Cancel   context.CancelFunc

	Configuration            config.HypermassConfig
	PublicationConfiguration config.PublicationConfiguration

	FolderPath string

	Disposer      payload_read_disposer.PayloadReadDisposerStrategy
	FileExtension string
}

// NewPublicationPoller create an active PublicationPoller and starts running it
func NewPublicationPoller(parentCtx context.Context, publicationConfig config.PublicationConfiguration, hypermassConfig config.HypermassConfig) (*PublicationPoller, error) {
	fmt.Println("Publishing to stream: " + publicationConfig.Key)
	ctx, cancel := context.WithCancel(parentCtx)

	folderPath := helpers.GetStreamPathFromConfig(publicationConfig.TargetDirectory)
	streamConfigFromService := publication_helpers.GetConfigurationForStream(hypermassConfig, publicationConfig.Key)
	directoryError := subscriptionhelpers.InitialiseAndCheckDirectory(folderPath)

	if directoryError != nil {
		log.Println(directoryError)
		log.Println("Unable to initialise directory")
		cancel()
		return nil, directoryError
	}

	subscription := PublicationPoller{
		StreamId:                 publicationConfig.Key,
		Ctx:                      ctx,
		Cancel:                   cancel,
		Configuration:            hypermassConfig,
		PublicationConfiguration: publicationConfig,
		FolderPath:               folderPath,
		Disposer:                 payload_read_disposer.GetPayloadReadDisposer(publicationConfig.DisposerType, publicationConfig.Key, publicationConfig.TargetDirectory),
		FileExtension:            streamConfigFromService.FileExtension,
	}

	go subscription.pollForFiles()

	return &subscription, nil
}

// pollForFiles poll for files - poll time can vary based on hypermass feedback. Ctx.Done() interrupts the poller wait.
func (s *PublicationPoller) pollForFiles() {
	for {
		nextDelayDuration := s.handleNextFilesInFolder()

		var nextDelay time.Duration
		if nextDelayDuration != nil {
			nextDelay = *nextDelayDuration
		} else {
			nextDelay = 5 * time.Second
		}

		select {
		case <-time.After(nextDelay):
			// Timer expired, continue polling
			continue
		case <-s.Ctx.Done():
			// Exit the poller
			log.Println("Publication poller stopped: ", s.StreamId)
			return
		}
	}
}

// handleNextFilesInFolder handles the next set of files from the polled folder, returning a wait interval in seconds if
// needed. Typically, this would be because a publishing rate limit has been reached and the server has advised a
// wait duration.
func (s *PublicationPoller) handleNextFilesInFolder() *time.Duration {

	filesToProcess, err := publication_helpers.FindNextFilesInFolder(s.FolderPath, s.FileExtension, MaxFileSize, MinFileAge)

	if err != nil {
		log.Printf("Error Scanning directory: %s %s \n", s.FolderPath, err)
		return nil
	}

	for _, entry := range filesToProcess {
		uploadOutcome, err := publication_helpers.PublishFileToStream(entry.Path, s.StreamId, s.Configuration.Token)

		if err != nil {
			var insufficientAllowanceError *app_errors.InsufficientAllowanceError
			var retryLaterError *app_errors.RetryLaterError

			//default retry behaviour for disconnections should be fairly frequent - e.g. recovering from network loss
			waitTime := time.Duration(60) * time.Second

			if errors.As(err, &insufficientAllowanceError) {
				//only retry for allowance changes every 5 minutes to prevent the service being overwhelmed
				waitTime = time.Duration(5) * time.Minute
				log.Println("unable to publish to stream "+s.StreamId+": ", err)
			} else if errors.As(err, &retryLaterError) {
				//server advises when to retry in this case
				waitTime = retryLaterError.RetryAfter
				fmt.Printf("Too soon to upload to stream " + s.StreamId + ", retry in " + waitTime.String())
			} else {
				log.Println("unable to publish to stream "+s.StreamId+": ", err)
			}

			log.Printf("Failed to upload payload file (%s) - retrying in: "+waitTime.String(), entry.Path)
			return &waitTime //break the loop
		}

		// upload was accepted
		fmt.Printf("Uploaded file (%s) to stream (%s) - payload id: %s \n", entry.Path, s.StreamId, uploadOutcome.PayloadId)

		for {
			err := s.Disposer.DisposeOfPayloadFile(entry.Path)

			if err == nil {
				break
			}

			fmt.Printf("Failed to clean up the uploaded file (%s), further uploads blocked. Please delete manually (poller will retry every 30 seconds): %s \n", entry.Path, err)
			time.Sleep(30 * time.Second)
		}
	}

	return nil
}
