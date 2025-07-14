package meshdump

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Server wraps the HTTP router and store.
type Server struct {
	store *Store
	mux   *http.ServeMux
}

func NewServer(store *Store) *Server {
	s := &Server{store: store, mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) Router() *http.ServeMux { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/api/telemetry/", s.handleTelemetry())
	s.mux.HandleFunc("/api/nodes", s.handleNodes)
	s.mux.HandleFunc("/", s.handleIndex)
}

func (s *Server) handleTelemetry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node := strings.TrimPrefix(r.URL.Path, "/api/telemetry/")
		if node == "" {
			http.Error(w, "missing node", http.StatusBadRequest)
			return
		}
		data := s.store.Get(node)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodes := s.store.Nodes()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

//go:embed web/index.html
var indexHTML string

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, indexHTML)
}
