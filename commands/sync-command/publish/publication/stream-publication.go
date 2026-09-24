package publication

import (
	"context"
	"errors"
	"fmt"
	"hypermass-cli/app_common"
	"hypermass-cli/app_errors"
	"hypermass-cli/commands/sync-command/helpers"
	"hypermass-cli/commands/sync-command/publish/payload_read_disposer"
	publication_helpers "hypermass-cli/commands/sync-command/publish/publication/publication-helpers"
	publication_status "hypermass-cli/commands/sync-command/publish/publication/publication-status"
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

	Auth                     config.HypermassAuth
	PublicationConfiguration config.PublicationConfiguration

	FolderPath string

	Disposer      payload_read_disposer.PayloadReadDisposerStrategy
	FileExtension string

	ReportingState publication_status.PublicationReportingState
}

// NewPublicationPoller create an active PublicationPoller and starts running it
func NewPublicationPoller(parentCtx context.Context, publicationConfig config.PublicationConfiguration, hypermassProfile config.HypermassProfile) (*PublicationPoller, error) {
	fmt.Println("Publishing to stream: " + publicationConfig.Key)
	ctx, cancel := context.WithCancel(parentCtx)

	folderPath := helpers.GetStreamPathFromConfig(publicationConfig.TargetDirectory)
	//an unreadable configuration still produces a poller. It starts in an error state, shows that
	//through the status command, and reads the configuration on a later pass once it becomes available
	streamConfigFromService, configError := publication_helpers.GetConfigurationForStream(
		hypermassProfile, publicationConfig.Key)

	directoryError := subscriptionhelpers.InitialiseAndCheckDirectory(folderPath)

	if directoryError != nil {
		log.Println(directoryError)
		log.Println("Unable to initialise directory")
		cancel()
		return nil, directoryError
	}

	var credentialsRejected *app_errors.CredentialsRejectedError
	if errors.As(configError, &credentialsRejected) {
		cancel()
		return nil, configError
	}

	publicationPoller := &PublicationPoller{
		StreamId:                 publicationConfig.Key,
		Ctx:                      ctx,
		Cancel:                   cancel,
		Auth:                     hypermassProfile.Auth,
		PublicationConfiguration: publicationConfig,
		FolderPath:               folderPath,
		Disposer:                 payload_read_disposer.GetPayloadReadDisposer(publicationConfig.DisposerType, publicationConfig.Key, publicationConfig.TargetDirectory),
		FileExtension:            streamConfigFromService.FileExtension,
		ReportingState:           initialStateFor(configError),
	}

	go publicationPoller.pollForFiles()

	return publicationPoller, nil
}

// initialStateFor gives a publication its starting state, which shows any configuration failure.
func initialStateFor(configError error) publication_status.PublicationReportingState {
	if configError == nil {
		return publication_status.NewInitialState()
	}

	return publication_status.NewErrorStatus(summaryOf(configError), retryAfterFor(configError), 0)
}

// noteAccountHealth updates the shared credential state from the outcome of a call.
//
// Refused credentials apply to the whole sync, and an accepted call shows they work again, so whichever
// stream makes the call keeps the state current for all of them.
func noteAccountHealth(err error) {
	var credentialsRejected *app_errors.CredentialsRejectedError

	if errors.As(err, &credentialsRejected) {
		app_common.RecordCredentialsRejected(credentialsRejected.Summary())
		return
	}

	if err == nil {
		app_common.RecordSuccessfulContact()
	}
}

// summaryOf returns the error's own description, for the status command to show against a publication.
func summaryOf(err error) string {
	var retryable app_errors.RetryableError

	if errors.As(err, &retryable) {
		return retryable.Summary()
	}

	return "waiting to retry"
}

// retryAfterFor returns the delay carried by the error, or the default when it carries none.
func retryAfterFor(err error) time.Duration {
	var retryable app_errors.RetryableError

	if errors.As(err, &retryable) {
		return retryable.RetryAfter()
	}

	return app_errors.DefaultRetryDelay
}

// pollForFiles poll for files - poll time can vary based on hypermass feedback. Ctx.Done() interrupts the poller wait.
func (s *PublicationPoller) pollForFiles() {
	for {
		nextDelayDuration := s.handleNextFilesInFolder()

		select {
		case <-time.After(*nextDelayDuration):
			// Timer expired, continue polling
			continue
		case <-s.Ctx.Done():
			// Exit the poller
			log.Println("Publication poller stopped: ", s.StreamId)
			return
		}
	}
}

// retryConfiguration reads a configuration that was unavailable earlier.
//
// It returns how long to wait when the configuration is still unavailable, and nil once it has been
// read and publishing can begin.
func (s *PublicationPoller) retryConfiguration() *time.Duration {
	streamConfig, err := publication_helpers.GetConfigurationForStream(
		config.HypermassProfile{Auth: s.Auth}, s.StreamId)

	noteAccountHealth(err)

	if err == nil {
		s.FileExtension = streamConfig.FileExtension
		return nil
	}

	wait := retryAfterFor(err)
	summary := summaryOf(err)

	log.Printf("Unable to read the configuration for %s: %v - retrying in %s", s.StreamId, err, wait)
	s.ReportingState = publication_status.NewErrorStatus(summary, wait, s.countRemainingFiles())

	return &wait
}

// handleNextFilesInFolder handles the next set of files from the polled folder, returning a wait interval in seconds if
// needed. Typically, this would be because a publishing rate limit has been reached and the server has advised a
// wait duration.
func (s *PublicationPoller) handleNextFilesInFolder() *time.Duration {

	//publishing needs the file type from the configuration, so read that before looking for files
	if s.FileExtension == "" {
		if wait := s.retryConfiguration(); wait != nil {
			return wait
		}
	}

	filesToProcess, err := publication_helpers.FindNextFilesInFolder(s.FolderPath, s.FileExtension, MaxFileSize, MinFileAge)
	fallbackWaitTime := time.Duration(20) * time.Second

	if err != nil {
		log.Printf("Error Scanning directory: %s %s \n", s.FolderPath, err)
		s.ReportingState = publication_status.NewErrorStatus("Error scanning directory", fallbackWaitTime, 0)
		return &fallbackWaitTime
	}

	for _, entry := range filesToProcess {
		s.ReportingState = publication_status.NewPublishingStatus(s.countRemainingFiles())

		uploadOutcome, err := publication_helpers.PublishFileToStream(entry.Path, s.StreamId, s.Auth.Token)

		if err != nil {
			var insufficientAllowanceError *app_errors.InsufficientAllowanceError
			var retryLaterError *app_errors.RetryLaterError

			//the error supplies the delay and the description, leaving only the choice of status to report
			waitTime := retryAfterFor(err)
			summary := summaryOf(err)

			if errors.As(err, &insufficientAllowanceError) {
				log.Println("unable to publish to stream "+s.StreamId+": ", err)
				s.ReportingState = publication_status.NewInsufficientAllowanceStatus(waitTime, s.countRemainingFiles())
			} else if errors.As(err, &retryLaterError) {
				fmt.Println("Too soon to upload to stream " + s.StreamId + ", retry in " + waitTime.String())
				s.ReportingState = publication_status.NewWaitingRateLimitedStatus(waitTime, s.countRemainingFiles())
			} else {
				log.Printf("unable to publish to stream %s: %v", s.StreamId, err)
				s.ReportingState = publication_status.NewErrorStatus(summary, waitTime, s.countRemainingFiles())
			}

			noteAccountHealth(err)

			log.Printf("Failed to upload payload file (%s) - retrying in: "+waitTime.String(), entry.Path)

			return &waitTime //break the loop
		}

		// upload was accepted
		noteAccountHealth(nil)
		fmt.Printf("Uploaded file (%s) to stream (%s) - payload id: %s \n", entry.Path, s.StreamId, uploadOutcome.PayloadId)

		for {
			err := s.Disposer.DisposeOfPayloadFile(entry.Path)

			if err == nil {
				break
			}

			fmt.Printf("Failed to clean up the uploaded file (%s), further uploads blocked. Please delete manually (poller will retry every 30 seconds): %s \n", entry.Path, err)
			sleepDuration := 30 * time.Second
			s.ReportingState = publication_status.NewDeletingFileFailedRetryingStatus(sleepDuration, s.countRemainingFiles())
			time.Sleep(sleepDuration)
		}
	}

	s.ReportingState = publication_status.NewPollingWaitStatus(fallbackWaitTime, s.countRemainingFiles())
	return &fallbackWaitTime
}

func (s *PublicationPoller) countRemainingFiles() int {
	updatedFiles, _ := publication_helpers.FindNextFilesInFolder(s.FolderPath, s.FileExtension, MaxFileSize, MinFileAge)
	return len(updatedFiles)
}
