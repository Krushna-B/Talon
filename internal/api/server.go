package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Krushna-B/talon/internal/config"
)

type Server struct {
	cfg config.Config
	log *slog.Logger //Only 1 logger resource shared via pointer
}
type HealthResponse struct {
	Status string `json:"status"`
	Mode   string `json:"mode"`
}

func NewServer(cfg config.Config, log *slog.Logger) *Server {
	server := Server{
		cfg: cfg,
		log: log,
	}
	return &server
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status: "ok",
		Mode:   s.cfg.Mode,
	})
}
