package kalshi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Market struct {
	Ticker       string    `json:"ticker"`
	EventTicker  string    `json:"event_ticker"`
	Status       string    `json:"status"`
	YesBid       string    `json:"yes_bid_dollars"`
	YesAsk       string    `json:"yes_ask_dollars"`
	NoBid        string    `json:"no_bid_dollars"`
	NoAsk        string    `json:"no_ask_dollars"`
	LastPrice    string    `json:"last_price_dollars"`
	Volume       string    `json:"volume_fp"`
	OpenInterest string    `json:"open_interest_fp"`
	CloseTime    time.Time `json:"close_time"`
}
type marketResponse struct {
	Markets []Market `json:"markets"`
	Cursor  string   `json:"cursor"`
}

func (c *Client) GetMarkets(ctx context.Context, status string, limit int) ([]Market, string, error) {
	// Build the URL.
	u, err := url.Parse(c.baseURL + "/markets")
	if err != nil {
		return nil, "", fmt.Errorf("parsing kalshi url: %w", err)
	}
	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	if status != "" {
		q.Set("status", status)
	}
	u.RawQuery = q.Encode()

	// Create the request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating http request: %w", err)
	}

	// Execute the request.
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("calling kalshi markets: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("kalshi markets returned status %d", res.StatusCode)
	}

	var out marketResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, "", fmt.Errorf("decoding kalshi markets: %w", err)
	}
	return out.Markets, out.Cursor, nil
}
