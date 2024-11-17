package dbfactory

import (
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

type Config = gormkitteh.Config

var Configs = configset.Set[Config]{
	Default: &Config{
		SlowQueryThreshold: 300 * time.Millisecond,
		Type:               gormkitteh.RequireExplicitDSNType,
	},
	Dev: &Config{
		SlowQueryThreshold: 200 * time.Millisecond, // gorm default
		DSN:                "sqlite:file:" + filepath.Join(xdg.DataHomeDir(), "autokitteh.sqlite"),
		AutoMigrate:        true,
	},
	Test: &Config{},
}

func New(z *zap.Logger, cfg *Config) (db.DB, error) {
	return dbgorm.New(z, cfg)
}
