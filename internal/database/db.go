package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the pgxpool.Pool
type DB struct {
	*pgxpool.Pool
}

// Config holds database configuration
type Config struct {
	URL            string
	MaxConns       int32
	MinConns       int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// New creates a new database connection pool
func New(cfg Config) (*DB, error) {
	// Parse connection string and create pool config
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Set connection pool settings
	poolConfig.MaxConns = cfg.MaxConns
	if poolConfig.MaxConns == 0 {
		poolConfig.MaxConns = 20 // default
	}

	poolConfig.MinConns = cfg.MinConns
	if poolConfig.MinConns == 0 {
		poolConfig.MinConns = 2 // default
	}

	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	if poolConfig.MaxConnLifetime == 0 {
		poolConfig.MaxConnLifetime = time.Hour
	}

	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	if poolConfig.MaxConnIdleTime == 0 {
		poolConfig.MaxConnIdleTime = 30 * time.Minute
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("Database connection pool closed")
	}
}

// Health checks the database connection health
func (db *DB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return db.Ping(ctx)
}
