package meshdump

import (
	"sync"
	"time"
)

type Telemetry struct {
	NodeID    string
	DataType  string
	Value     float64
	Timestamp time.Time
}

type Store struct {
	mu   sync.Mutex
	data map[string][]Telemetry
}

func NewStore() *Store {
	return &Store{data: make(map[string][]Telemetry)}
}

func (s *Store) Add(t Telemetry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[t.NodeID] = append(s.data[t.NodeID], t)
}

func (s *Store) Get(nodeID string) []Telemetry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Telemetry(nil), s.data[nodeID]...)
}

func (s *Store) All() map[string][]Telemetry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string][]Telemetry, len(s.data))
	for k, v := range s.data {
		out[k] = append([]Telemetry(nil), v...)
	}
	return out
}
