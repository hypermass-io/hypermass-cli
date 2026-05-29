package subscription_status

import "time"

type SubscriptionReportingState struct {
	Status       string
	Description  string
	WaitDuration time.Duration
	WaitingUntil time.Time

	LastActivity time.Time
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
