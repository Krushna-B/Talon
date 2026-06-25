package config

import (
	"errors"
	"fmt"
	"os"
)

// Trading Types
const (
	ModePaper = "paper"
	ModeLive  = "live"
)

// Config holds the runtime settings shared by all Talon workers.
type Config struct {
	DatabaseURL   string
	Mode          string
	HTTPAddr      string
	AdminToken    string
	KalshiBaseURL string

	KalshiWSURL   string
	KalshiKeyID   string
	KalshiKeyPath string
}

// Load reads configuration from environment variables.
//
//	DATABASE_URL is required.
//	Mode default is PaperMode
//	HTTPAddr default is :8080
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Mode:          getenv("MODE", ModePaper),
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		AdminToken:    os.Getenv("ADMIN_TOKEN"),
		KalshiBaseURL: getenv("KALSHI_BASE_URL", "https://external-api.demo.kalshi.co/trade-api/v2"),
		KalshiWSURL:   getenv("KALSHI_WS_URL", "wss://external-api-ws.demo.kalshi.co/trade-api/ws/v2"),
		KalshiKeyID:   os.Getenv("KALSHI_KEY_ID"),
		KalshiKeyPath: getenv("KALSHI_KEY_PATH", "private.pem"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL not set")
	}
	if cfg.AdminToken == "" {
		return Config{}, errors.New("ADMIN_TOKEN not set")
	}
	if cfg.Mode != ModePaper && cfg.Mode != ModeLive {
		return Config{}, fmt.Errorf("invalid MODE %q: must be %q or %q", cfg.Mode, ModePaper, ModeLive)
	}

	return cfg, nil
}

// getenv returns the value of key, or fallback if key is unset or empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
