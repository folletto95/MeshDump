package meshdump

import (
	"encoding/json"
	"log"
	"os"
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
	file string
}

func NewStore(path string) *Store {
	s := &Store{data: make(map[string][]Telemetry), file: path}
	if path != "" {
		_ = s.load()
	}
	return s
}

func (s *Store) Add(t Telemetry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("store: add node=%s type=%s value=%f", t.NodeID, t.DataType, t.Value)
	s.data[t.NodeID] = append(s.data[t.NodeID], t)
	if s.file != "" {
		_ = s.saveLocked()
	}
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

func (s *Store) Nodes() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	nodes := make([]string, 0, len(s.data))
	for k := range s.data {
		nodes = append(nodes, k)
	}
	return nodes
}

func (s *Store) saveLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, b, 0644)
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, &s.data)
}
