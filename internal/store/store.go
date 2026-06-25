package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUnknownKey is returned when a system_state key does not exist.
var ErrUnknownKey = errors.New("unknown system_state key")

// Store is the application's gateway to the database.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a connection pool and verifies the database is reachable.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close releases all connections in the pool.
func (s *Store) Close() {
	s.pool.Close()
}

// Ping verifies the database is reachable.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// GetSystemState returns the value of a system_state row by key.
func (s *Store) GetSystemState(ctx context.Context, key string) (string, error) {
	var value string
	err := s.pool.QueryRow(ctx, "SELECT value FROM system_state WHERE key = $1", key).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("getting system state %q: %w", key, err)
	}
	return value, nil
}

// SetSystemState sets the values of specifc system_state row by key.
func (s *Store) SetSystemState(ctx context.Context, key string, val string) error {
	tag, err := s.pool.Exec(ctx, "UPDATE system_state SET value = $1, updated_at = now() WHERE key = $2", val, key)
	if err != nil {
		return fmt.Errorf("unable to update %q state: %w", key, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("setting system state %q: %w", key, ErrUnknownKey)
	}

	return nil
}

// Order is the data needed to write-ahead a new order row. Plain fields so
// store stays decoupled from the kalshi package.
type Order struct {
	OrderID    string
	Ticker     string
	Side       string
	Action     string
	Count      int
	LimitPrice float64
}

// InsertPending records an order as 'pending' BEFORE the venue call, so a
// crash mid-submit leaves a recoverable trace (write-ahead).
func (s *Store) InsertPending(ctx context.Context, o Order) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO orders (order_id, ticker, side, action, count, limit_price, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending')`,
		o.OrderID, o.Ticker, o.Side, o.Action, o.Count, o.LimitPrice)
	if err != nil {
		return fmt.Errorf("inserting pending order: %w", err)
	}
	return nil
}

// MarkResting fills in the venue's id and flips status to 'resting' once
// the venue acks the order.
func (s *Store) MarkResting(ctx context.Context, orderID, venueOrderID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE orders SET venue_order_id = $1, status = 'resting', updated_at = now() WHERE order_id = $2`,
		venueOrderID, orderID)
	if err != nil {
		return fmt.Errorf("marking order resting: %w", err)
	}
	return nil
}

// ListSystemState returns the values of all system_state row by key.
func (s *Store) ListSystemState(ctx context.Context) (map[string]string, error) {
	rows, err := s.pool.Query(ctx, "SELECT key, value FROM system_state")
	if err != nil {
		return nil, fmt.Errorf("listing system state: %w", err)
	}
	defer rows.Close()
	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scanning system state: %w", err)
		}
		out[k] = v
	}
	return out, rows.Err()
}
