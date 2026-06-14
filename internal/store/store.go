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
		return fmt.Errorf("unknown system_state key %q", key)
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
