package meshdump

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Server wraps the HTTP router and store.
type Server struct {
	store *Store
	mux   *mux.Router
}

func NewServer(store *Store) *Server {
	s := &Server{store: store, mux: mux.NewRouter()}
	s.routes()
	return s
}

func (s *Server) Router() *mux.Router { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/api/telemetry/{node}", s.handleTelemetry()).
		Methods(http.MethodGet)
}

func (s *Server) handleTelemetry() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		node := vars["node"]
		data := s.store.Get(node)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}
