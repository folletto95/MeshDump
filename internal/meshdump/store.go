package meshdump

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Telemetry struct {
	NodeID    string
	DataType  string
	Value     float64
	Timestamp time.Time
}

// NodeInfo describes a node by ID with optional names.
type NodeInfo struct {
	ID        string
	LongName  string
	ShortName string
}

// Store keeps telemetry and node information in memory and persists it to an
// optional SQLite database using the `sqlite3` command line utility.
type Store struct {
	mu    sync.Mutex
	data  map[string][]Telemetry
	nodes map[string]NodeInfo
	file  string
	debug bool
}

// NewStore initializes the store. When path is non-empty a SQLite database is
// created (if necessary) and used for persistence.
func NewStore(path string) *Store {

	s := &Store{
		data:  make(map[string][]Telemetry),
		nodes: make(map[string]NodeInfo),
		file:  path,
		debug: os.Getenv("DEBUG") != "" && os.Getenv("DEBUG") != "0",
	}
	if path != "" {
		_ = s.initDB()
		_ = s.load()
	}
	if s.debug {
		log.Printf("store debug enabled")
	}
	return s
}

// Add stores a telemetry entry in memory and on disk.
func (s *Store) Add(t Telemetry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("store: add node=%s type=%s value=%f", t.NodeID, t.DataType, t.Value)
	s.data[t.NodeID] = append(s.data[t.NodeID], t)
	if _, ok := s.nodes[t.NodeID]; !ok {
		s.nodes[t.NodeID] = NodeInfo{ID: t.NodeID}
		if s.file != "" {
			sql := fmt.Sprintf("INSERT OR IGNORE INTO nodes (node_id, long_name, short_name) VALUES (%q,'','');", t.NodeID)
			_ = exec.Command("sqlite3", s.file, sql).Run()
		}
		if s.debug {
			log.Printf("debug: discovered node %s", t.NodeID)
		}
	}
	if s.file != "" {
		ts := t.Timestamp.Format(time.RFC3339Nano)
		sql := fmt.Sprintf("INSERT INTO telemetry (node_id, data_type, value, timestamp) VALUES (%q,%q,%f,%q);", t.NodeID, t.DataType, t.Value, ts)
		_ = exec.Command("sqlite3", s.file, sql).Run()
	}
}

// Get returns telemetry for the given node ID.
func (s *Store) Get(nodeID string) []Telemetry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Telemetry(nil), s.data[nodeID]...)
}

// All returns a copy of the telemetry map.
func (s *Store) All() map[string][]Telemetry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string][]Telemetry, len(s.data))
	for k, v := range s.data {
		out[k] = append([]Telemetry(nil), v...)
	}
	return out
}

// Nodes returns the list of known nodes with associated metadata.
func (s *Store) Nodes() []NodeInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]NodeInfo, 0, len(s.nodes))
	for _, n := range s.nodes {
		out = append(out, n)
	}
	return out
}

// SetNodeInfo stores metadata about a node.
func (s *Store) SetNodeInfo(info NodeInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[info.ID] = info
	if s.file != "" {
		sql := fmt.Sprintf("INSERT OR REPLACE INTO nodes (node_id, long_name, short_name) VALUES (%q,%q,%q);", info.ID, info.LongName, info.ShortName)
		_ = exec.Command("sqlite3", s.file, sql).Run()
	}
	if s.debug {
		log.Printf("debug: node info updated %+v", info)

	}
}

// Node retrieves metadata for the given node ID. If not present, the returned
// boolean is false.
func (s *Store) Node(id string) (NodeInfo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.nodes[id]
	return n, ok
}

// initDB creates the required tables if they do not exist.
func (s *Store) initDB() error {
	schema := `CREATE TABLE IF NOT EXISTS telemetry (
    node_id TEXT,
    data_type TEXT,
    value REAL,
    timestamp TEXT
);
CREATE TABLE IF NOT EXISTS nodes (
    node_id TEXT PRIMARY KEY,
    long_name TEXT,
    short_name TEXT
);`
	return exec.Command("sqlite3", s.file, schema).Run()
}

// load repopulates the in-memory store from the SQLite database.
func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.file == "" {
		return nil
	}
	// load nodes
	out, err := exec.Command("sqlite3", "-json", s.file, "SELECT node_id, long_name, short_name FROM nodes;").Output()
	if err == nil && len(out) > 0 {
		var rows []struct {
			ID        string `json:"node_id"`
			LongName  string `json:"long_name"`
			ShortName string `json:"short_name"`
		}
		if err := json.Unmarshal(out, &rows); err == nil {
			for _, r := range rows {
				s.nodes[r.ID] = NodeInfo{ID: r.ID, LongName: r.LongName, ShortName: r.ShortName}
			}
		}
	}

	if s.debug && len(s.nodes) > 0 {
		log.Printf("debug: loaded nodes %+v", s.nodes)
	}

	// load telemetry
	out, err = exec.Command("sqlite3", "-json", s.file, "SELECT node_id, data_type, value, timestamp FROM telemetry;").Output()
	if err == nil && len(out) > 0 {
		var rows []struct {
			NodeID    string  `json:"node_id"`
			DataType  string  `json:"data_type"`
			Value     float64 `json:"value"`
			Timestamp string  `json:"timestamp"`
		}
		if err := json.Unmarshal(out, &rows); err == nil {
			for _, r := range rows {
				ts, _ := time.Parse(time.RFC3339Nano, r.Timestamp)
				s.data[r.NodeID] = append(s.data[r.NodeID], Telemetry{
					NodeID:    r.NodeID,
					DataType:  r.DataType,
					Value:     r.Value,
					Timestamp: ts,
				})
			}
		}
	}
	return nil
}
