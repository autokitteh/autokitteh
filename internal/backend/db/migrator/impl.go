package migrator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pressly/goose/v3"

	"go.autokitteh.dev/autokitteh/migrations"
)

type Migrator struct {
	db            *sql.DB
	dialect       string
	DbVersion     int64
	migrationsDir string
}

func initGoose(client *sql.DB, dialect string) (ver int64, err error) {
	goose.SetBaseFS(migrations.Migrations)

	if err = goose.SetDialect(dialect); err != nil {
		return
	}

	if ver, err = goose.EnsureDBVersion(client); err != nil {
		err = fmt.Errorf("failed to ensure DB version: %w", err)
	}

	return
}

func (m *Migrator) Migrate(ctx context.Context) error {
	if err := goose.UpContext(ctx, m.db, m.migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

func (m *Migrator) MigrationRequired(ctx context.Context) (bool, error) {
	requiredMigrations, err := goose.CollectMigrations(m.migrationsDir, m.DbVersion, int64((1<<63)-1))
	if err != nil && !errors.Is(err, goose.ErrNoMigrationFiles) {
		return false, err
	}

	return len(requiredMigrations) > 0, nil
}
