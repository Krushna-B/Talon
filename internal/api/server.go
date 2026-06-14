package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Krushna-B/talon/internal/config"
	"github.com/Krushna-B/talon/internal/store"
)

type Server struct {
	cfg   config.Config
	log   *slog.Logger
	store *store.Store
}

type HealthResponse struct {
	Status string `json:"status"`
	Mode   string `json:"mode"`
	DB     string `json:"db"`
}

func NewServer(cfg config.Config, log *slog.Logger, store *store.Store) *Server {
	return &Server{
		cfg:   cfg,
		log:   log,
		store: store,
	}
}

// HTTP Server mux
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /admin/state", s.handleGetState)
	mux.HandleFunc("POST /admin/state", s.handleSetState)
	return mux
}

// API Handlers
// Returns Health Response json
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status: "ok",
		Mode:   s.cfg.Mode,
		DB:     "ok",
	}
	status := http.StatusOK

	if err := s.store.Ping(r.Context()); err != nil {
		s.log.Error("health check: database unreachable", "err", err)
		resp.Status = "degraded"
		resp.DB = "down"
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.log.Error("health check: encoding response", "err", err)
	}
}

// Returns all systesm states as a json
func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	state, err := s.store.ListSystemState(r.Context())
	if err != nil {
		s.log.Error("listing system state", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		s.log.Error("encoding system state response", "err", err)
	}
}

type setStateRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Server) handleSetState(w http.ResponseWriter, r *http.Request) {
	var req setStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Key == "" || req.Value == "" {
		http.Error(w, "key and value are required", http.StatusBadRequest)
		return
	}

	if err := s.store.SetSystemState(r.Context(), req.Key, req.Value); err != nil {
		if errors.Is(err, store.ErrUnknownKey) {
			http.Error(w, "unknown key", http.StatusBadRequest)
			return
		}
		s.log.Error("setting system state", "key", req.Key, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	state, err := s.store.ListSystemState(r.Context())
	if err != nil {
		s.log.Error("listing system state", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		s.log.Error("encoding system state response", "err", err)
	}
}
