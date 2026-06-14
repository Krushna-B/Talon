package kalshi

import (
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	log        *slog.Logger
}

func New(baseURL string, log *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		log:        log,
	}
}
