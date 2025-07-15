package meshdump

import (
	"database/sql"
	"log"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Telemetry struct {
	NodeID    string
	DataType  string
	Value     float64
	Timestamp time.Time
}

// NodeInfo describes a node by ID with optional names.
type NodeInfo struct {
	ID        string `json:"id"`
	LongName  string `json:"long_name"`
	ShortName string `json:"short_name"`
	Firmware  string `json:"firmware"`
}

// Store keeps telemetry and node information in memory and persists it to an
// optional SQLite database using the `sqlite3` command line utility.
type Store struct {
	mu    sync.Mutex
	data  map[string][]Telemetry
	nodes map[string]NodeInfo
	file  string
	debug bool
	db    *sql.DB
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
		db, err := sql.Open("sqlite", path)
		if err == nil {
			s.db = db
			_ = s.initDB()
			_ = s.load()
		} else {
			log.Printf("store: %v", err)
		}
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
		if s.db != nil {
			_, _ = s.db.Exec("INSERT OR IGNORE INTO nodes (node_id, long_name, short_name, firmware) VALUES (?, '', '', '')", t.NodeID)
		}
		if s.debug {
			log.Printf("debug: discovered node %s", t.NodeID)
		}
	}
	if s.db != nil {
		ts := t.Timestamp.Format(time.RFC3339Nano)
		_, _ = s.db.Exec("INSERT INTO telemetry (node_id, data_type, value, timestamp) VALUES (?, ?, ?, ?)", t.NodeID, t.DataType, t.Value, ts)
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
	if s.db != nil {
		_, _ = s.db.Exec("INSERT OR REPLACE INTO nodes (node_id, long_name, short_name, firmware) VALUES (?, ?, ?, ?)", info.ID, info.LongName, info.ShortName, info.Firmware)
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
	schema := `CREATE TABLE IF NOT EXISTS nodes (
    node_id TEXT PRIMARY KEY,
    long_name TEXT,
    short_name TEXT,
    firmware TEXT
);
CREATE TABLE IF NOT EXISTS telemetry (
    node_id TEXT NOT NULL,
    data_type TEXT,
    value REAL,
    timestamp TEXT,
    FOREIGN KEY(node_id) REFERENCES nodes(node_id)
);
CREATE INDEX IF NOT EXISTS idx_telemetry_node_id ON telemetry(node_id);`
	if s.db == nil {
		return nil
	}
	_, err := s.db.Exec(schema)
	return err
}

// load repopulates the in-memory store from the SQLite database.
func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	// load nodes
	rows, err := s.db.Query("SELECT node_id, long_name, short_name, firmware FROM nodes")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, long, short, fw string
			if err := rows.Scan(&id, &long, &short, &fw); err == nil {
				s.nodes[id] = NodeInfo{ID: id, LongName: long, ShortName: short, Firmware: fw}
			}
		}
	}

	if s.debug && len(s.nodes) > 0 {
		log.Printf("debug: loaded nodes %+v", s.nodes)
	}

	// load telemetry
	trows, err := s.db.Query("SELECT node_id, data_type, value, timestamp FROM telemetry")
	if err == nil {
		defer trows.Close()
		for trows.Next() {
			var nodeID, dtype, tsStr string
			var val float64
			if err := trows.Scan(&nodeID, &dtype, &val, &tsStr); err == nil {
				ts, _ := time.Parse(time.RFC3339Nano, tsStr)
				s.data[nodeID] = append(s.data[nodeID], Telemetry{
					NodeID:    nodeID,
					DataType:  dtype,
					Value:     val,
					Timestamp: ts,
				})
			}
		}
	}
	return nil
}

// Close closes the underlying database if open.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
