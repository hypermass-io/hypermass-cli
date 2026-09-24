package subscription

import (
	"errors"
	"fmt"
	"hypermass-cli/app_common"
	"hypermass-cli/app_errors"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	subscription_status "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-status"
	"log"
	"math/rand"
	"sync"
	"time"
)

type SubscriptionPollers struct {
	mu   sync.Mutex
	data map[string]*Subscription
	WG   sync.WaitGroup
}

func NewSubscriptionPollers() *SubscriptionPollers {
	return &SubscriptionPollers{
		data: make(map[string]*Subscription),
	}
}

func (s *SubscriptionPollers) Store(key string, value *Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value.RequestRestart = s.handleRequestRestart

	s.data[key] = value

	//adds a block worker for the pollers context, which itself may have many child workers
	s.WG.Go(func() {
		<-value.Ctx.Done()
	})
}

func (s *SubscriptionPollers) Load(key string) (*Subscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	return value, ok
}

// Snapshot returns a shallow copy snapshot of the map
func (s *SubscriptionPollers) Snapshot() map[string]*Subscription {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := make(map[string]*Subscription, len(s.data))
	for key, subscription := range s.data {
		snapshot[key] = subscription
	}

	return snapshot
}

func (s *SubscriptionPollers) ResetToPayloadId(streamId string, payloadId string) (*Subscription, error) {
	oldSub, exists := s.Load(streamId)
	if !exists {
		return nil, fmt.Errorf("stream %s not found", streamId)
	}

	// Temporary worker to keep the pollers WG alive for the duration of the reset.
	s.WG.Add(1)

	log.Printf("Resetting stream %s. Purging queue...", streamId)
	oldSub.Cancel()
	log.Printf("⏳ Waiting for %s cleanup...", streamId)
	oldSub.ProcessorsWG.Wait()

	err := subscriptionhelpers.WriteLastPayloadId(oldSub.FolderPath, payloadId)
	if err != nil {
		return nil, fmt.Errorf("failed to reset state on disk: %w", err)
	}

	newSub, err := NewSubscription(oldSub.ParentCtx, oldSub.SubscriptionConfiguration, oldSub.Auth,
		time.Duration(0), subscription_status.NewInitialState(time.Duration(0)))
	if err != nil {
		return nil, err
	}

	s.Store(streamId, newSub)

	log.Printf("✅ Stream %s successfully reset to %s", streamId, payloadId)

	// Mark the temporary worker as done
	s.WG.Done()

	return newSub, nil
}

func (s *SubscriptionPollers) handleRequestRestart(streamId string, reason error) {
	oldSub, exists := s.Load(streamId)
	if !exists {
		log.Printf("unable to handle stopped subscription %s - not in SubscriptionPollers store", streamId)
		return
	}

	duration := determineDurationForError(reason)
	noteAccountHealth(reason)

	// Add a temporary worker so the main WG cannot run dry while we are switching subscriptions.
	s.WG.Go(func() {
		log.Printf("Connection lost for stream %s", oldSub.StreamId)

		oldSub.Cancel()
		log.Printf("⏳ Waiting for %s cleanup...", streamId)
		oldSub.ProcessorsWG.Wait()

		log.Printf("Subscriber for %s stopped, reason: %s", streamId, reason)

		log.Println("Retrying connection to " + oldSub.StreamId + " in " + duration.String() + "...")

		waitingState := subscription_status.NewWaitingAfterErrorState(duration, summaryOf(reason))

		newSub, err := NewSubscription(oldSub.ParentCtx, oldSub.SubscriptionConfiguration, oldSub.Auth,
			duration, waitingState)
		if err != nil {
			log.Printf("unable to handle stopped subscription %s - failed to create replacement: %v", streamId, err)
			return
		}

		s.Store(streamId, newSub)
	})

	return
}

func noteAccountOk() {
	noteAccountHealth(nil)
}

// noteAccountHealth updates the shared credential state from the outcome of a call. The publication
// side has the same helper, for the same reason.
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

// summaryOf returns the error's own description, for the status command to show against a subscription.
func summaryOf(err error) string {
	var retryable app_errors.RetryableError

	if errors.As(err, &retryable) {
		return retryable.Summary()
	}

	if err != nil {
		return "waiting to reconnect"
	}

	return "reconnecting"
}

func determineDurationForError(err error) time.Duration {
	var retryable app_errors.RetryableError

	//the error carries its own delay, so this function needs no knowledge of the kinds of error. The
	//jitter is added here because spreading out reconnections is this loop's concern
	if errors.As(err, &retryable) {
		return retryable.RetryAfter() + standardJitter()
	}

	if err != nil {
		return app_errors.DefaultRetryDelay + standardJitter()
	}

	//a clean exit that still wants restarting, so recover promptly
	return (time.Duration(15) * time.Second) + highJitter()
}

func standardJitter() time.Duration {
	const jitterMaxSeconds = 10
	return time.Duration(rand.Intn(jitterMaxSeconds)) * time.Second
}

func highJitter() time.Duration {
	const jitterMaxSeconds = 20
	return time.Duration(rand.Intn(jitterMaxSeconds)) * time.Second
}
