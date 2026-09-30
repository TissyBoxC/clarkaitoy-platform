// Package database owns PostgreSQL connectivity.
package database

import "context"

// Store is a placeholder for database-backed repositories.
type Store struct{}

// Close releases database resources.
func (s *Store) Close(_ context.Context) error {
	return nil
}
