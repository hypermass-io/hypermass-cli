package subscription

import (
	"errors"
	"fmt"
	"hypermass-cli/app_errors"
	subscriptionhelpers "hypermass-cli/commands/sync-command/subscribe/subscription/subscription-helpers"
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

	newSub, err := NewSubscription(oldSub.ParentCtx, oldSub.SubscriptionConfiguration, oldSub.Auth, time.Duration(0))
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

	// Add a temporary worker so the main WG cannot run dry while we are switching subscriptions.
	s.WG.Go(func() {
		log.Printf("Connection lost for stream %s", oldSub.StreamId)

		oldSub.Cancel()
		log.Printf("⏳ Waiting for %s cleanup...", streamId)
		oldSub.ProcessorsWG.Wait()

		log.Printf("Subscriber for %s stopped, reason: %s", streamId, reason)

		log.Println("Retrying connection to " + oldSub.StreamId + " in " + duration.String() + "...")
		newSub, err := NewSubscription(oldSub.ParentCtx, oldSub.SubscriptionConfiguration, oldSub.Auth, duration)
		if err != nil {
			log.Printf("unable to handle stopped subscription %s - failed to create replacement: %w", streamId, err)
			return
		}

		s.Store(streamId, newSub)
	})

	return
}

func determineDurationForError(err error) time.Duration {
	var insufficientAllowanceError *app_errors.InsufficientAllowanceError
	var connectionLostError *app_errors.ConnectionLostError

	var duration time.Duration

	if errors.As(err, &insufficientAllowanceError) {
		//only poll for allowance changes every 5 minutes to prevent the service being overwhelmed
		duration = time.Duration(5) * time.Minute
	} else if errors.As(err, &connectionLostError) {
		//connection loss should re-try rapidly, but with higher jitter to avoid stampede
		duration = (time.Duration(10) * time.Second) + highJitter()
	} else if err != nil {
		//this is a fallback, normally not expecting this to happen
		duration = time.Duration(60)*time.Second + standardJitter()
	} else {
		//default poll behaviour catches all other types of errors.
		// Fairly frequent (for speedy recovery) without being aggressive
		duration = (time.Duration(15) * time.Second) + highJitter()
	}

	return duration
}

func standardJitter() time.Duration {
	const jitterMaxSeconds = 10
	return time.Duration(rand.Intn(jitterMaxSeconds)) * time.Second
}

func highJitter() time.Duration {
	const jitterMaxSeconds = 20
	return time.Duration(rand.Intn(jitterMaxSeconds)) * time.Second
}
