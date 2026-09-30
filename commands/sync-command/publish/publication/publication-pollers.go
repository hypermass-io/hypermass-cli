package publication

import (
	"hypermass-cli/config"
	"sort"
	"sync"
)

// PublicationPollers a repository of pollers
type PublicationPollers struct {
	mu     sync.Mutex
	data   map[string]*PublicationPoller
	WG     sync.WaitGroup
	closed bool
}

func NewPublicationPollers() *PublicationPollers {
	return &PublicationPollers{
		data: make(map[string]*PublicationPoller),
	}
}

// Store holds the poller for the key, and the WaitGroup waits on it until it stops. Once closed it cancels the
// poller instead.
func (s *PublicationPollers) Store(key string, value *PublicationPoller) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		value.Cancel()
		return
	}
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

// Close stops the pollers taking on publications, so the WaitGroup can be waited on at shutdown. Anything stored
// after Close is cancelled instead.
func (s *PublicationPollers) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
}

// Take removes and returns the poller held for the key.
func (s *PublicationPollers) Take(key string) (*PublicationPoller, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	delete(s.data, key)
	return value, ok
}

// RunningConfigurations returns the configuration of each started publication, ordered by key. Failed entries
// are left out, so a reload retries them.
func (s *PublicationPollers) RunningConfigurations() []config.PublicationConfiguration {
	s.mu.Lock()
	defer s.mu.Unlock()

	var running []config.PublicationConfiguration
	for _, poller := range s.data {
		if poller.StartError == nil {
			running = append(running, poller.PublicationConfiguration)
		}
	}

	sort.Slice(running, func(i, j int) bool { return running[i].Key < running[j].Key })
	return running
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
