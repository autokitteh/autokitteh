package gormkitteh

import (
	"errors"
	"fmt"

	// mattn/sqlite does not play well without cgo. this is problematic for
	// cross compiling.
	// See https://github.com/go-gorm/gorm/issues/4101.
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open a DB according to given configuration.
// Open calls cfg.Explicit() is called before evaluating it.
func Open(cfg *Config, f func(*gorm.Config)) (*gorm.DB, error) {
	var dialector gorm.Dialector

	cfg, err := cfg.Explicit()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	switch cfg.Type {
	case "sqlite":
		dialector = sqlite.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		return nil, ErrUnknownType
	}

	if dialector == nil {
		return nil, errors.New("no dialector")
	}

	gormCfg := gorm.Config{TranslateError: true}
	if f != nil {
		f(&gormCfg)
	}

	db, err := gorm.Open(dialector, &gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err) // TODO: Explicitly specify this in an open error?
	}

	if cfg.Debug {
		db = db.Debug()
	}

	return db, nil
}
