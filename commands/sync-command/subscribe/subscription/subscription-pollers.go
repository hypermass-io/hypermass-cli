package subscription

import (
	"fmt"
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

	value.OnStop = s.handleStoppedSubscriber

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
	return newSub, nil
}

func (s *SubscriptionPollers) handleStoppedSubscriber(streamId string, reason error) {
	oldSub, exists := s.Load(streamId)
	if !exists {
		log.Printf("unable to handle stopped subscription %s - not in SubscriptionPollers store", streamId)
		return
	}

	log.Printf("Subscriber for %s stopped, reason: %s", streamId, reason)
	log.Printf("Resetting stream %s. Purging queue...", streamId)

	newSub, err := NewSubscription(oldSub.ParentCtx, oldSub.SubscriptionConfiguration, oldSub.Auth, secondsWithJitter(60, 5))
	if err != nil {
		log.Printf("unable to handle stopped subscription %s - failed to create replacement: %w", streamId, err)
		return
	}

	s.Store(streamId, newSub)

	log.Printf("✅ Stream %s successfully recreated", streamId)
	return
}

func secondsWithJitter(secondsDuration int, jitterMaxSeconds int) time.Duration {
	jitter := time.Duration(rand.Intn(jitterMaxSeconds)) * time.Second
	return time.Duration(secondsDuration)*time.Second + jitter
}
