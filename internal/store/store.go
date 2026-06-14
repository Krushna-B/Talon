package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

// GetSystemState returns the value of a system_state row by key.
func (s *Store) GetSystemState(ctx context.Context, key string) (string, error) {
	var value string
	err := s.pool.QueryRow(ctx, "SELECT value FROM system_state WHERE key = $1", key).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("getting system state %q: %w", key, err)
	}
	return value, nil
}
