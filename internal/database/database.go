package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver registered as "pgx"
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

	// Tune connection pool for a small API server
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &DB{db}, nil
}

// Migrate runs the SQL migration file to create tables if they don't exist.
// Safe to call on every startup — uses CREATE TABLE IF NOT EXISTS.
func (db *DB) Migrate() error {
	query, err := os.ReadFile("internal/database/migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}
	if _, err = db.Exec(string(query)); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}
	return nil
}
