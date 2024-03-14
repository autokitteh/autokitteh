package dbfactory

import (
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/db"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
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
