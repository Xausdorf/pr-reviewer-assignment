package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	log "github.com/sirupsen/logrus"
)

func RunMigrations(ctx context.Context, databaseURL, migrationsDir string, logger *log.Logger) error {
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		logger.WithError(err).Error("Failed to open database for migrations")
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.WithError(err).Error("Failed to create migrate postgres driver")
		return fmt.Errorf("postgres driver: %w", err)
	}

	sourceURL := "file://" + migrationsDir
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		logger.WithError(err).WithField("source", sourceURL).Error("Failed to create migrate instance")
		return fmt.Errorf("create migrate: %w", err)
	}
	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.WithError(err).Error("Failed running migrations")
		return fmt.Errorf("migrate up: %w", err)
	}

	logger.WithField("source", sourceURL).Info("Migrations applied (no change or up completed)")
	return nil
}
