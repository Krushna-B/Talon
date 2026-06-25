package kalshi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
)

const wsPath = "/trade-api/ws/v2"

// Json sent to subscribe to certain channels
type subscribeCMD struct {
	ID     int          `json:"id"`
	Cmd    string       `json:"cmd"`
	Params subscribeAns `json:"params"`
}

type subscribeAns struct {
	Channels      []string `json:"channels"`
	MarketTickers []string `json:"market_tickers"`
}

// wsEnvelope is stage one of the decode: every Kalshi ws message shares this
// shape. msg stays as raw bytes so we can decode it differently per type
type wsEnvelope struct {
	Type string          `json:"type"`
	Msg  json.RawMessage `json:"msg"`
}

// MarketTick is the normalized snapshot
// Prices are in dollars (0–1) and equal the market's implied probability
// (yes_ask 0.41 → market thinks ~41% likely). Bid/ask are the YES side values
// NO is the complement (1 − yes)
type MarketTick struct {
	Ticker   string
	YesBid   float64
	YesAsk   float64
	Last     float64 // last trade price, dollars
	Volume   float64
	TSMillis int64
}

// tickerMsg is stage two: the concrete shape of a "ticker" message's msg field
type tickerMsg struct {
	MarketTicker string  `json:"market_ticker"`
	YesBid       float64 `json:"yes_bid_dollars,string"`
	YesAsk       float64 `json:"yes_ask_dollars,string"`
	Price        float64 `json:"price_dollars,string"`
	Volume       float64 `json:"volume_fp,string"`
	TSMillis     int64   `json:"ts_ms"`
}

func (c *Client) StreamTickers(ctx context.Context, wsUrl string, signer *Signer, tickers []string, out chan<- MarketTick) error {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sig, err := signer.Sign(ts, http.MethodGet, wsPath)
	if err != nil {
		return fmt.Errorf("signing handshake: %w", err)
	}

	header := http.Header{}
	header.Set("KALSHI-ACCESS-KEY", signer.KeyID())
	header.Set("KALSHI-ACCESS-SIGNATURE", sig)
	header.Set("KALSHI-ACCESS-TIMESTAMP", ts)

	conn, resp, err := websocket.Dial(ctx, wsUrl, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			c.log.Error("ws handshake rejected",
				"status", resp.StatusCode,
				"body", string(body))
		}
		return fmt.Errorf("dialing kalshi ws: %w", err)
	}
	defer conn.Close(websocket.StatusInternalError, "client closing") //Clean up

	sub := subscribeCMD{
		ID:  1,
		Cmd: "subscribe",
		Params: subscribeAns{
			Channels:      []string{"ticker"},
			MarketTickers: tickers,
		},
	}
	if err := writeJSON(ctx, conn, sub); err != nil {
		return fmt.Errorf("subscribing: %w", err)
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("reading ws message: %w", err)
		}

		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			c.log.Warn("undecodable ws message", "raw", string(data))
			continue
		}

		switch env.Type {
		case "ticker":
			var m tickerMsg
			if err := json.Unmarshal(env.Msg, &m); err != nil {
				c.log.Warn("bad ticker payload", "err", err, "raw", string(env.Msg))
				continue
			}
			tick := MarketTick{
				Ticker:   m.MarketTicker,
				YesBid:   m.YesBid,
				YesAsk:   m.YesAsk,
				Last:     m.Price,
				Volume:   m.Volume,
				TSMillis: m.TSMillis,
			}
			select {
			case out <- tick:
			default:
				// consumer is behind: drop this stale tick rather than
				// block the socket read and back up the connection
			}
		case "subscribed":
			c.log.Info("subscription confirmed", "raw", string(env.Msg))
		case "error":
			c.log.Error("ws error message", "raw", string(env.Msg))
		default:
			c.log.Debug("unhandled ws message", "type", env.Type, "raw", string(env.Msg))
		}
	}
}

func writeJSON(ctx context.Context, conn *websocket.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling: %w", err)
	}
	return conn.Write(ctx, websocket.MessageText, data)
}
