package publication_status

import (
	"fmt"
	"time"
)

type PublicationReportingState struct {
	Status           string
	Description      string
	QueueSize        int
	NextPoll         time.Time
	PollWaitDuration time.Duration
}

func NewInitialState() PublicationReportingState {
	return PublicationReportingState{
		Status:      "Initialising",
		Description: "initialising...",
		QueueSize:   0,
	}
}

func NewWaitingRateLimitedStatus(pollWaitDuration time.Duration, queued int) PublicationReportingState {
	nextPollTime := time.Now().Add(pollWaitDuration)

	return PublicationReportingState{
		Status:           "Rate-Limited",
		Description:      fmt.Sprintf("waiting until %s (delay of %s) to publish", nextPollTime.Format(time.RFC3339), pollWaitDuration.String()),
		NextPoll:         nextPollTime,
		PollWaitDuration: pollWaitDuration,
		QueueSize:        queued,
	}
}

func NewPollingWaitStatus(pollWaitDuration time.Duration, queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:           "Polling",
		Description:      "Waiting for next poll",
		NextPoll:         time.Now().Add(pollWaitDuration),
		PollWaitDuration: pollWaitDuration,
		QueueSize:        queued,
	}
}

func NewPublishingStatus(queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:      "Publishing",
		Description: "Publishing file (%s queued, including this one)",
		QueueSize:   queued,
	}
}

func NewErrorStatus(errorText string, pollWaitTime time.Duration, queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:           "Error",
		Description:      errorText,
		NextPoll:         time.Now().Add(pollWaitTime),
		PollWaitDuration: pollWaitTime,
		QueueSize:        queued,
	}
}

func NewInsufficientAllowanceStatus(pollWaitTime time.Duration, queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:           "Insufficient-Allowance",
		Description:      "Insufficient allowance - polling for status change",
		NextPoll:         time.Now().Add(pollWaitTime),
		PollWaitDuration: pollWaitTime,
		QueueSize:        queued,
	}
}

func NewDeletingFileFailedRetryingStatus(pollWaitTime time.Duration, queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:           "Delete-File-Failed",
		Description:      "Unable to delete file, retrying",
		NextPoll:         time.Now().Add(pollWaitTime),
		PollWaitDuration: pollWaitTime,
		QueueSize:        queued,
	}
}
