package meshdump

import (
	"embed"
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
	s.mux.HandleFunc("/api/nodeinfo/", s.handleNodeInfo())
	s.mux.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.FS(libFS))))
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
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodes := s.store.Nodes()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleNodeInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/nodeinfo/")
		if id == "" {
			http.Error(w, "missing node id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var info NodeInfo
			if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if info.ID == "" {
				info.ID = id
			}
			s.store.SetNodeInfo(info)
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			info, ok := s.store.Node(id)
			w.Header().Set("Content-Type", "application/json")
			if !ok {
				if _, err := w.Write([]byte("{}")); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
			if err := json.NewEncoder(w).Encode(info); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

//go:embed web/index.html
var indexHTML string

//go:embed web/lib/*
var libFS embed.FS

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if _, err := io.WriteString(w, indexHTML); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
