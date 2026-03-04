package database

import (
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending migrations using a dedicated connection.
// m.Close() would close the underlying *sql.DB, so we use a separate connection
// and leave the app's connection untouched.
func RunMigrations() error {
	migrateDB, err := NewPostgresConnection()
	if err != nil {
		return err
	}
	defer migrateDB.Close()

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrations source: %w", err)
	}

	dbDriver, err := postgres.WithInstance(migrateDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("Migrations: already up to date")
	} else {
		log.Println("Migrations: applied successfully")
	}

	return nil
}
