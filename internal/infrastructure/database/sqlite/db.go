package sqlite

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	sqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func OpenDatabase(dataDir string) (*sql.DB, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	dbPath := filepath.Join(dataDir, "infinita.db")
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := runMigrations(db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("close database after migration failure: %v", closeErr)
		}
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	if err := os.Chmod(dbPath, 0o600); err != nil && !os.IsNotExist(err) {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("close database after chmod failure: %v", closeErr)
		}
		return nil, fmt.Errorf("chmod db: %w", err)
	}
	return db, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("sqlite driver: %w", err)
	}
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("iofs source: %w", err)
	}
	migrator, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
