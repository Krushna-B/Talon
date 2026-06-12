package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Krushna-B/talon/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "talon:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	slog.Info("starting talon", "mode", cfg.Mode, "addr", cfg.HTTPAddr)
	return nil
}
