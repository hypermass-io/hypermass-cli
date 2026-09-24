package subscription_status

import "time"

type SubscriptionReportingState struct {
	Status       string
	Description  string
	WaitDuration time.Duration
	WaitingUntil time.Time

	LastActivity time.Time

	// LastError describes why this subscription is waiting, when it is waiting because of a failure.
	// The status command shows it alongside the time remaining, so a stalled subscription says what
	// stalled it.
	LastError string
}

func NewInitialState(initialWait time.Duration) SubscriptionReportingState {

	if initialWait == 0 {
		return SubscriptionReportingState{
			Status:       "Initialising",
			Description:  "Initialising...",
			LastActivity: time.Now(),
		}
	} else {
		return SubscriptionReportingState{
			Status:       "Waiting-To-Connect",
			Description:  "Waiting to connect (" + initialWait.String() + ")",
			WaitDuration: initialWait,
			WaitingUntil: time.Now().Add(initialWait),
			LastActivity: time.Now(),
		}
	}
}

// NewWaitingAfterErrorState describes a subscription that is backing off after a failure.
func NewWaitingAfterErrorState(wait time.Duration, summary string) SubscriptionReportingState {
	//the description is the cause only, since the status table has its own column for time remaining
	return SubscriptionReportingState{
		Status:       "Waiting-To-Reconnect",
		Description:  summary,
		WaitDuration: wait,
		WaitingUntil: time.Now().Add(wait),
		LastActivity: time.Now(),
		LastError:    summary,
	}
}

func NewStoppedState() SubscriptionReportingState {
	return SubscriptionReportingState{
		Status:       "Stopped",
		Description:  "Stopped",
		LastActivity: time.Now(),
	}
}

func NewRestartingState() SubscriptionReportingState {
	return SubscriptionReportingState{
		Status:       "Restarting",
		Description:  "Restarting",
		LastActivity: time.Now(),
	}
}

func NewConnectingState() SubscriptionReportingState {
	return SubscriptionReportingState{
		Status:       "Connecting",
		Description:  "Connecting",
		LastActivity: time.Now(),
	}
}

func NewConnectedActivityState() SubscriptionReportingState {
	return SubscriptionReportingState{
		Status:       "Connected",
		Description:  "Connected",
		LastActivity: time.Now(),
	}
}
