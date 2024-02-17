package dbfactory

import (
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/configset"
	"go.autokitteh.dev/autokitteh/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/backend/internal/db"
	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm"
)

type Config = gormkitteh.Config

var Configs = configset.Set[Config]{
	Default: &Config{
		SlowQueryThreshold: 300 * time.Millisecond,
	},
}

func New(z *zap.Logger, cfg *Config) (db.DB, error) {
	return dbgorm.New(z, cfg)
}
