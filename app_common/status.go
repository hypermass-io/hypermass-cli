package app_common

import "time"

type SubscriptionStatus struct {
	StreamId     string `json:"streamId"`
	Status       string
	Description  string
	WaitDuration time.Duration
	WaitingUntil time.Time
	LastActivity time.Time

	LastPayloadId  string `json:"lastPayloadId"`
	StartPoint     string `json:"startPoint"`
	FetchQueueSize int    `json:"fetchQueueSize"`
}

type PublicationStatus struct {
	StreamId         string        `json:"streamId"`
	Status           string        `json:"status"`
	Description      string        `json:"description"`
	PollWaitDuration time.Duration `json:"pollWaitDuration"`
	NextPoll         time.Time     `json:"nextPoll"`
	QueueSize        int           `json:"queueSize"`
}

type StatusReport struct {
	Subscriptions []SubscriptionStatus `json:"subscriptions"`
	Publications  []PublicationStatus  `json:"publicationStatus"`
}
