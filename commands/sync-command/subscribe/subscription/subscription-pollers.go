package subscription

import (
	"errors"
	"fmt"
	"hypermass-cli/app_common"
	"hypermass-cli/app_errors"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
	subscription_status "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-status"
	"hypermass-cli/config"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type SubscriptionPollers struct {
	mu     sync.Mutex
	data   map[string]*Subscription
	WG     sync.WaitGroup
	closed bool
}

func NewSubscriptionPollers() *SubscriptionPollers {
	return &SubscriptionPollers{
		data: make(map[string]*Subscription),
	}
}

func (s *SubscriptionPollers) Store(key string, value *Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.adopt(key, value)
}

// Replace stores a replacement subscription only while the old subscription is still the one held for the key.
// Protects against the 'old' subscription being reloaded, and the (now stale) replacement regressing the reload.
// Also avoids having to lock _all subscriptions_ as we wait for the old subscription to finish.
// A refused replacement is cancelled.
func (s *SubscriptionPollers) Replace(key string, old *Subscription, replacement *Subscription) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data[key] != old {
		replacement.Cancel()
		return false
	}

	return s.adopt(key, replacement)
}

// Close stops the pollers taking on subscriptions, so the WaitGroup can be waited on at shutdown. Anything stored
// or replaced after Close is cancelled instead.
func (s *SubscriptionPollers) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

// adopt makes the subscription the one held for the key: its restart requests come back here, and the
// pollers' WaitGroup waits on it until it stops. Once closed it cancels the subscription instead, returning false.
// Callers must hold s.mu.
func (s *SubscriptionPollers) adopt(key string, value *Subscription) bool {
	if s.closed {
		value.Cancel()
		return false
	}

	value.RequestRestart = s.handleRequestRestart

	s.data[key] = value

	//adds a block worker that holds until the stream is stopped and its processors have finished
	s.WG.Go(func() {
		<-value.Ctx.Done()
		value.ProcessorsWG.Wait()
	})

	return true
}

func (s *SubscriptionPollers) Load(key string) (*Subscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	return value, ok
}

// Take removes and returns the subscription held for the key, so any pending replacement for it is refused.
func (s *SubscriptionPollers) Take(key string) (*Subscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	delete(s.data, key)
	return value, ok
}

// RunningConfigurations returns the configuration of each started subscription, ordered by key. Failed entries
// are left out, so a reload retries them.
func (s *SubscriptionPollers) RunningConfigurations() []config.SubscriptionConfiguration {
	s.mu.Lock()
	defer s.mu.Unlock()

	var running []config.SubscriptionConfiguration
	for _, subscription := range s.data {
		if subscription.StartError == nil {
			running = append(running, subscription.SubscriptionConfiguration)
		}
	}

	sort.Slice(running, func(i, j int) bool { return running[i].Key < running[j].Key })
	return running
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

	if !s.Replace(streamId, oldSub, newSub) {
		return nil, fmt.Errorf("stream %s was changed while resetting", streamId)
	}

	log.Printf("✅ Stream %s successfully reset to %s", streamId, payloadId)

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

	// Stops oldSub and replaces it once its processors have finished. Asynchronous because the caller is one of
	// those processors, so waiting for them here would never finish.
	go func() {
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

		if !s.Replace(streamId, oldSub, newSub) {
			log.Printf("Subscription %s was replaced or removed while reconnecting, reconnection abandoned", streamId)
		}
	}()

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
