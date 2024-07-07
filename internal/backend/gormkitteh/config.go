// TODO: Make public?
package gormkitteh

import (
	"errors"
	"strings"
	"time"
)

const RequireExplicitDSNType = "require"

type Config struct {
	// If `Type` is empty, `DSN` must be in the form of "type:actual_dsn".
	// If both `Type` and `DSN` are empty, `Type` will be considered "sqlite"
	// with an empty DSN, which will make sqlite use a temporary database
	// (see https://www.sqlite.org/inmemorydb.html).
	// If `Type` is "require" (RequireExplicitDSNType), `DSN` must be specified.
	Type  string `koanf:"type"`
	DSN   string `koanf:"dsn"`
	Debug bool   `koanf:"debug"`

	SlowQueryThreshold time.Duration `koanf:"slow_query_threshold"`

	// If true, DB migrations will run automatically.
	// If false, the server will fail to start if a migration is required,
	// and the user has to run 'ak server migrate' explicitly.
	AutoMigrate bool `koanf:"auto_migrate"`

	Ownership string `koanf:"ownership"`
}

func (c Config) Explicit() (*Config, error) {
	if c.Type == RequireExplicitDSNType {
		if c.DSN == "" {
			return nil, errors.New("db config must be specified")
		}

		c.Type = ""
	}

	if c.Type == "" {
		if c.DSN == "" {
			// With empty DSN, assume type is "sqlite". Will make sqlite
			// use a temporary database.
			c.Type = "sqlite"
			// For in-memory, cached is required for transactions to work.
			// Not using "?cache=shared", for auto-cleanup between system tests.
			c.DSN = "file::memory:?cache=shared"
		} else {
			// This should make it easier to just specify the type in the dsn
			// in a single env var.
			var ok bool
			c.Type, c.DSN, ok = strings.Cut(c.DSN, ":")
			if !ok {
				return nil, ErrInvalidDSN
			}
		}
	}

	if c.Ownership != "none" {
		c.Ownership = "users"
	}

	return &c, nil
}
