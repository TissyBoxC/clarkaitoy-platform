// Package database owns PostgreSQL connectivity.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store owns the PostgreSQL connection pool shared by database-backed
// repositories. Callers must not close the pool themselves.
type Store struct {
	pool *pgxpool.Pool
}

// Open creates and verifies a PostgreSQL connection pool.
//
// The DSN must be supplied by secret-backed runtime configuration. Open never
// falls back to a shared database account.
func Open(ctx context.Context, dsn string) (*Store, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database DSN: %w", err)
	}
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 15 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Pool returns the initialized PostgreSQL pool.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Close releases database resources.
func (s *Store) Close(_ context.Context) error {
	if s == nil || s.pool == nil {
		return nil
	}
	s.pool.Close()
	s.pool = nil
	return nil
}
