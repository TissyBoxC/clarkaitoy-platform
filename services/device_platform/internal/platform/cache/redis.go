// Package cache owns Redis connectivity.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache owns the Redis client used for short-lived sessions, locks, and
// idempotency records. It is safe for concurrent use.
type Cache struct {
	client *redis.Client
}

// Open creates and verifies a Redis client.
func Open(ctx context.Context, address string, password string, database int) (*Cache, error) {
	if address == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     password,
		DB:           database,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Cache{client: client}, nil
}

// Client returns the initialized Redis client.
func (c *Cache) Client() *redis.Client {
	return c.client
}

// Close releases Redis resources.
func (c *Cache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
