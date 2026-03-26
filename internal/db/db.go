package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is satisfied by both *pgxpool.Pool and pgx.Tx, allowing helpers
// to be called within or outside a transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var DB *pgxpool.Pool

func Init() error {
	ctx := context.Background()
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	// Conservative defaults for a self-hosted app
	config.MaxConns = 25
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	// Simple retry loop for startup (e.g., waiting for Docker DB)
	var lastErr error
	for i := 0; i < 10; i++ {
		DB, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			// Test the connection
			if err = DB.Ping(ctx); err == nil {
				return nil
			}
			DB.Close()
		}
		lastErr = err
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("failed to connect to database after retries: %w", lastErr)
}
