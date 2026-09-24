package publication_status

import (
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
		Description: "initialising",
		QueueSize:   0,
	}
}

func NewWaitingRateLimitedStatus(pollWaitDuration time.Duration, queued int) PublicationReportingState {
	nextPollTime := time.Now().Add(pollWaitDuration)

	return PublicationReportingState{
		Status: "Rate-Limited",
		//the cause only, since the status table has its own columns for the next poll and time remaining
		Description:      "publishing too fast, waiting",
		NextPoll:         nextPollTime,
		PollWaitDuration: pollWaitDuration,
		QueueSize:        queued,
	}
}

func NewPollingWaitStatus(pollWaitDuration time.Duration, queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:           "Polling",
		Description:      "waiting for the next file poll",
		NextPoll:         time.Now().Add(pollWaitDuration),
		PollWaitDuration: pollWaitDuration,
		QueueSize:        queued,
	}
}

func NewPublishingStatus(queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:      "Publishing",
		Description: "publishing a file",
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
		Description:      "insufficient allowance",
		NextPoll:         time.Now().Add(pollWaitTime),
		PollWaitDuration: pollWaitTime,
		QueueSize:        queued,
	}
}

func NewDeletingFileFailedRetryingStatus(pollWaitTime time.Duration, queued int) PublicationReportingState {
	return PublicationReportingState{
		Status:           "Delete-File-Failed",
		Description:      "could not delete a published file",
		NextPoll:         time.Now().Add(pollWaitTime),
		PollWaitDuration: pollWaitTime,
		QueueSize:        queued,
	}
}
