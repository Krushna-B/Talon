package main

import (
	"context"
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
	"github.com/Krushna-B/talon/internal/store"
	"github.com/Krushna-B/talon/internal/strategy"
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

	// consumer: run each tick through the strategy, log any intents
	go func() {
		for tick := range ticks {
			for _, intent := range strat.OnTick(tick) {
				slog.Info("intent",
					"ticker", intent.Ticker, "side", intent.Side,
					"action", intent.Action, "count", intent.Count,
					"limit", intent.LimitPrice)
			}
		}
	}()

	go func() {
		tickers := []string{}
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
