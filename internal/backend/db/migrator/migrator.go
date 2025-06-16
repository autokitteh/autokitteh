//go:build !enterprise
// +build !enterprise

package migrator

import (
	"database/sql"
)

func NewMigrator(db *sql.DB, dialect string) (*Migrator, error) {
	dbVersion, err := initGoose(db, dialect)
	if err != nil {
		return nil, err
	}

	return &Migrator{db: db, dialect: dialect, DbVersion: dbVersion, migrationsDir: dialect}, nil
}
