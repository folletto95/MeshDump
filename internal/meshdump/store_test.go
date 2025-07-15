package meshdump

import (
	"path/filepath"
	"testing"
	"time"
)

func TestStorePersistence(t *testing.T) {
	dir := t.TempDir()
	db := filepath.Join(dir, "data.db")

	s := NewStore(db)
	defer s.Close()

	tel := Telemetry{NodeID: "n1", DataType: "temp", Value: 42.0, Timestamp: time.Now()}
	s.Add(tel)
	s.SetNodeInfo(NodeInfo{ID: "n1", LongName: "node1"})

	// reopen store
	s.Close()
	s2 := NewStore(db)
	defer s2.Close()

	info, ok := s2.Node("n1")
	if !ok || info.LongName != "node1" {
		t.Fatalf("expected node info preserved")
	}

	data := s2.Get("n1")
	if len(data) != 1 {
		t.Fatalf("expected one telemetry entry, got %d", len(data))
	}
	if data[0].DataType != "temp" || data[0].Value != 42.0 {
		t.Fatalf("unexpected telemetry %+v", data[0])
	}
}
