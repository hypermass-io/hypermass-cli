package publication

import "sync"

// PublicationPollers a repository of pollers
type PublicationPollers struct {
	mu   sync.Mutex
	data map[string]*PublicationPoller
	WG   sync.WaitGroup
}

func NewPublicationPollers() *PublicationPollers {
	return &PublicationPollers{
		data: make(map[string]*PublicationPoller),
	}
}

func (s *PublicationPollers) Store(key string, value *PublicationPoller) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value

	//adds a block worker that holds until the stream is stopped and its processors have finished
	s.WG.Go(func() {
		<-value.Ctx.Done()
		value.ProcessorsWG.Wait()
	})
}

func (s *PublicationPollers) Load(key string) (*PublicationPoller, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	return value, ok
}

// Snapshot returns a shallow copy snapshot of the map
func (s *PublicationPollers) Snapshot() map[string]*PublicationPoller {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := make(map[string]*PublicationPoller, len(s.data))

	for key, value := range s.data {
		snapshot[key] = value
	}

	return snapshot
}
