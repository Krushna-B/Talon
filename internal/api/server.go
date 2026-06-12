package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Krushna-B/talon/internal/config"
)

type Server struct {
	cfg config.Config
	log *slog.Logger //Only 1 logger resource shared via pointer
}

func New(cfg config.Config, log *slog.Logger) *Server {
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
	fmt.Fprintln(w, "ok")
}
