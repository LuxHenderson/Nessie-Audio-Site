package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationFiles embed.FS

// RunMigrations applies all pending database migrations
// This is called automatically on server startup
func RunMigrations(db *sql.DB) error {
	log.Println("🔄 Checking for database migrations...")

	// Create driver for migrations
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	// Create source from embedded files
	sourceDriver, err := iofs.New(migrationFiles, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	// Create migrator
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("get migration version: %w", err)
	}

	if dirty {
		log.Printf("⚠️  WARNING: Database is in dirty state at version %d", version)
		log.Println("   This usually means a previous migration failed partway through.")
		log.Println("   You may need to manually fix the migration or force the version.")
		return fmt.Errorf("database in dirty state at version %d", version)
	}

	if errors.Is(err, migrate.ErrNilVersion) {
		log.Println("📝 No migrations applied yet, running initial migration...")
	} else {
		log.Printf("📊 Current migration version: %d", version)
	}

	// Run migrations
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("✅ Database schema is up to date")
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	// Get new version
	newVersion, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("get new version: %w", err)
	}

	log.Printf("✅ Successfully applied migrations (now at version %d)", newVersion)
	return nil
}

// GetCurrentVersion returns the current migration version
// Useful for debugging and monitoring
func GetCurrentVersion(db *sql.DB) (uint, bool, error) {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationFiles, ".")
	if err != nil {
		return 0, false, fmt.Errorf("create migration source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", driver)
	if err != nil {
		return 0, false, fmt.Errorf("create migrator: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, err
	}

	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}

	return version, dirty, nil
}
