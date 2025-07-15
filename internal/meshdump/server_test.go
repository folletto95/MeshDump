package meshdump

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer() (*Server, *Store) {
	st := NewStore("")
	srv := NewServer(st)
	return srv, st
}

func TestTelemetryHandler(t *testing.T) {
	srv, st := newTestServer()
	tel := Telemetry{NodeID: "n1", DataType: "temperature", Value: 10, Timestamp: time.Unix(1700000001, 0)}
	st.Add(tel)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/telemetry/n1", nil)
	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []Telemetry
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 || got[0].NodeID != tel.NodeID ||
		got[0].DataType != tel.DataType ||
		got[0].Value != tel.Value ||
		!got[0].Timestamp.Equal(tel.Timestamp) {
		t.Errorf("unexpected body: %v", got)
	}

	// missing node
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/telemetry/", nil)
	srv.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing node, got %d", rr.Code)
	}
}

func TestNodesHandler(t *testing.T) {
	srv, st := newTestServer()
	st.Add(exampleTelemetry)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/nodes", nil)
	srv.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var nodes []NodeInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &nodes); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(nodes) != 1 || nodes[0].ID != exampleTelemetry.NodeID {
		t.Errorf("unexpected nodes: %v", nodes)
	}
}

func TestNodeInfoHandler(t *testing.T) {
	srv, st := newTestServer()
	st.Add(exampleTelemetry)

	// POST update node info
	info := NodeInfo{LongName: "Tester", Firmware: "2.0"}
	body, _ := json.Marshal(info)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/nodeinfo/"+exampleTelemetry.NodeID, bytes.NewReader(body))
	srv.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	stored, ok := st.Node(exampleTelemetry.NodeID)
	if !ok || stored.LongName != "Tester" || stored.Firmware != "2.0" {
		t.Errorf("node info not stored: %+v", stored)
	}

	// GET existing node info
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/nodeinfo/"+exampleTelemetry.NodeID, nil)
	srv.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got NodeInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != exampleTelemetry.NodeID || got.LongName != "Tester" {
		t.Errorf("unexpected node info: %+v", got)
	}
}
