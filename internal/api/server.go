package api

import (
	"encoding/json"
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

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

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
