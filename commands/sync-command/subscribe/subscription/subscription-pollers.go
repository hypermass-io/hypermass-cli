package subscription

import "sync"

type SubscriptionPollers struct {
	mu   sync.Mutex
	data map[string]Subscription
	WG   sync.WaitGroup
}

func NewSubscriptionPollers() *SubscriptionPollers {
	return &SubscriptionPollers{
		data: make(map[string]Subscription),
	}
}

func (s *SubscriptionPollers) Store(key string, value Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value

	//adds a block worker for the pollers context, which itself may have many child workers
	s.WG.Go(func() {
		<-value.Ctx.Done()
	})
}

func (s *SubscriptionPollers) Load(key string) (Subscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	return value, ok
}
