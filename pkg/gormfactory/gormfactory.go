package gormfactory

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var cache = make(map[string]*gorm.DB, 16)

type Config struct {
	Type  string `envconfig:"TYPE" default:"sqlite" json:"type"`
	DSN   string `envconfig:"DSN" default:"file::memory:?cache=shared" json:"dsn"`
	Debug bool   `envconfig:"DEBUG" json:"debug"`
}

func (c Config) IsZero() bool { return c.Type == "" && c.DSN == "" }

func OpenInMem() (*gorm.DB, error) { return Open(Config{}) }

func MustOpenInMem() *gorm.DB {
	db, err := OpenInMem()
	if err != nil {
		panic(err)
	}
	return db
}

func Open(cfg Config) (*gorm.DB, error) {
	cacheKey := fmt.Sprintf("%s.%s", cfg.Type, cfg.DSN)

	db := cache[cacheKey]

	if db == nil {
		var dialector gorm.Dialector

		switch cfg.Type {
		case "", "sqlite":
			dsn := cfg.DSN
			if dsn == "" {
				dsn = "file::memory:?cache=shared"
			}
			dialector = sqlite.Open(dsn)
		case "postgres":
			dialector = postgres.Open(cfg.DSN)
		default:
			return nil, fmt.Errorf("unknown type: %s", cfg.Type)
		}

		l := logger.New(
			// TODO: use L
			log.Default(),
			logger.Config{
				SlowThreshold:             time.Second,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		).LogMode(logger.Warn)

		var err error
		if db, err = gorm.Open(dialector, &gorm.Config{
			Logger: l,
		}); err != nil {
			return nil, err
		}

		cache[cacheKey] = db
	}

	if cfg.Debug {
		db = db.Debug()
	}

	return db, nil
}
