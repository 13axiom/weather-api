package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps sql.DB to add project-specific methods.
type DB struct {
	*sql.DB
}

// New connects to PostgreSQL and verifies the connection.
func New(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err = db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return &DB{db}, nil
}

// Migrate runs all SQL migration files in order.
// Safe to call on every startup — all statements use IF NOT EXISTS.
func (db *DB) Migrate() error {
	migrations := []string{
		"internal/database/migrations/001_init.sql",
		"internal/database/migrations/002_air_quality.sql",
	}
	for _, path := range migrations {
		query, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		if _, err = db.Exec(string(query)); err != nil {
			return fmt.Errorf("run migration %s: %w", path, err)
		}
	}
	return nil
}
