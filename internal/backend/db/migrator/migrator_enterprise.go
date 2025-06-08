//go:build enterprise
// +build enterprise

package migrator

import (
	"database/sql"
	"fmt"
)

func NewMigrator(db *sql.DB, dialect string) (*Migrator, error) {
	if dialect != "postgres" {
		return nil, fmt.Errorf("unsupported dialect for enterprise: %s", dialect)
	}

	dbVersion, err := initGoose(db, dialect)
	if err != nil {
		return nil, err
	}

	return &Migrator{db: db, dialect: dialect, DbVersion: dbVersion, migrationsDir: "postgres/enterprise"}, nil
}
