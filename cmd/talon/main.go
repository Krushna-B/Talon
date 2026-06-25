package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Krushna-B/talon/internal/api"
	"github.com/Krushna-B/talon/internal/config"
	"github.com/Krushna-B/talon/internal/kalshi"
	"github.com/Krushna-B/talon/internal/paper"
	"github.com/Krushna-B/talon/internal/store"
	"github.com/Krushna-B/talon/internal/strategy"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "talon:", err)
		os.Exit(1)
	}
}

// newOrderID returns a random client-generated order id, used both as our
// primary key and the venue's idempotency token.
func newOrderID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting talon", "mode", cfg.Mode, "addr", cfg.HTTPAddr)

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to store: %w", err)
	}
	defer st.Close()

	//Build Kalshi event
	signer, err := kalshi.NewSigner(cfg.KalshiKeyID, cfg.KalshiKeyPath)
	if err != nil {
		return fmt.Errorf("building kalshi signer: %w", err)
	}
	kc := kalshi.New(cfg.KalshiBaseURL, slog.Default())

	srv := api.NewServer(cfg, slog.Default(), st)

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      srv.Routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	//GO Routines
	errCh := make(chan error, 2)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()
	ticks := make(chan kalshi.MarketTick, 1024)

	strat := strategy.CheapYes{MaxAsk: 0.30}
	broker := paper.New()

	// consumer: tick → strategy → intent → write-ahead → broker → mark resting
	go func() {
		for tick := range ticks {
			for _, intent := range strat.OnTick(tick) {
				orderID := newOrderID()

				if err := st.InsertPending(ctx, store.Order{
					OrderID: orderID, Ticker: intent.Ticker,
					Side: string(intent.Side), Action: string(intent.Action),
					Count: intent.Count, LimitPrice: intent.LimitPrice,
				}); err != nil {
					slog.Error("persisting pending order", "err", err)
					continue
				}

				res, err := broker.PlaceOrder(ctx, intent, orderID)
				if err != nil {
					slog.Error("placing order", "err", err) // row stays 'pending'
					continue
				}

				if err := st.MarkResting(ctx, orderID, res.VenueOrderID); err != nil {
					slog.Error("marking order resting", "err", err)
					continue
				}
				slog.Info("order placed",
					"order_id", orderID, "venue_id", res.VenueOrderID,
					"status", res.Status, "ticker", intent.Ticker)
			}
		}
	}()

	go func() {
		// Small throwaway list of currently-active cheap markets for testing
		// the order path. Swap freely; widen once the risk gate exists.
		tickers := []string{
			"KXBTCD-26JUN2617-T63999.99",
			"KXWCSCORE-26JUN25TURUSA-TUR0USA4",
			"KXMLBSPREAD-26JUN251545ATHSF-ATH4",
			"KXSPXFOMC-26JUL29-T2.75",
			"KXHIGHTSEA-26JUN25-T73",
		}
		errCh <- kc.StreamTickers(ctx, cfg.KalshiWSURL, signer, tickers, ticks)
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := httpServer.Shutdown(shutdownCtx)
		if err != nil {
			return fmt.Errorf("server: %w", err)
		}
	}
	return nil
}
