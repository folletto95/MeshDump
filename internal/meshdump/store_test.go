package meshdump

import (
	"testing"
	"time"
)

var exampleTelemetry = Telemetry{
	NodeID:    "node1",
	DataType:  "temperature",
	Value:     42.5,
	Timestamp: time.Unix(1700000000, 0),
}

func TestStoreAddGet(t *testing.T) {
	s := NewStore("")
	s.Add(exampleTelemetry)

	got := s.Get(exampleTelemetry.NodeID)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].NodeID != exampleTelemetry.NodeID ||
		got[0].DataType != exampleTelemetry.DataType ||
		got[0].Value != exampleTelemetry.Value ||
		!got[0].Timestamp.Equal(exampleTelemetry.Timestamp) {
		t.Errorf("unexpected telemetry: %+v", got[0])
	}

	if _, ok := s.Node(exampleTelemetry.NodeID); !ok {
		t.Errorf("node info should be created on Add")
	}
}

func TestStoreSetNodeInfo(t *testing.T) {
	s := NewStore("")
	info := NodeInfo{ID: "node2", LongName: "Node Two", ShortName: "n2", Firmware: "1.0"}
	s.SetNodeInfo(info)

	got, ok := s.Node("node2")
	if !ok {
		t.Fatalf("node info not found")
	}
	if got != info {
		t.Errorf("expected %+v, got %+v", info, got)
	}

	list := s.Nodes()
	if len(list) != 1 || list[0] != info {
		t.Errorf("unexpected nodes list: %+v", list)

	}
}
