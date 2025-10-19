package database

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrator handles database migrations
type Migrator struct {
	pool *pgxpool.Pool
}

// NewMigrator creates a new migrator
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{pool: pool}
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort up migrations
	var upMigrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}
	sort.Strings(upMigrations)

	// Run each migration
	for _, migrationFile := range upMigrations {
		// Extract version from filename (e.g., "001" from "001_create_users_table.up.sql")
		version := strings.Split(migrationFile, "_")[0]

		// Check if migration has already been applied
		applied, err := m.isMigrationApplied(ctx, version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if applied {
			log.Printf("Migration %s already applied, skipping", migrationFile)
			continue
		}

		// Read migration file
		content, err := migrationsFS.ReadFile("migrations/" + migrationFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
		}

		// Execute migration
		log.Printf("Applying migration: %s", migrationFile)
		if _, err := m.pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationFile, err)
		}

		// Record migration
		if err := m.recordMigration(ctx, version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migrationFile, err)
		}

		log.Printf("Successfully applied migration: %s", migrationFile)
	}

	log.Println("All migrations applied successfully")
	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down(ctx context.Context) error {
	// Get the last applied migration
	var version string
	err := m.pool.QueryRow(ctx, `
		SELECT version FROM schema_migrations
		ORDER BY version DESC
		LIMIT 1
	`).Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	// Find the corresponding down migration file
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var downFile string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), version) && strings.HasSuffix(entry.Name(), ".down.sql") {
			downFile = entry.Name()
			break
		}
	}

	if downFile == "" {
		return fmt.Errorf("down migration file not found for version %s", version)
	}

	// Read migration file
	content, err := migrationsFS.ReadFile("migrations/" + downFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", downFile, err)
	}

	// Execute migration
	log.Printf("Rolling back migration: %s", downFile)
	if _, err := m.pool.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", downFile, err)
	}

	// Remove migration record
	if _, err := m.pool.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	log.Printf("Successfully rolled back migration: %s", downFile)
	return nil
}

// createMigrationsTable creates the schema_migrations table
func (m *Migrator) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW() NOT NULL
		)
	`
	_, err := m.pool.Exec(ctx, query)
	return err
}

// isMigrationApplied checks if a migration has been applied
func (m *Migrator) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	var count int
	err := m.pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// recordMigration records that a migration has been applied
func (m *Migrator) recordMigration(ctx context.Context, version string) error {
	_, err := m.pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
	return err
}
