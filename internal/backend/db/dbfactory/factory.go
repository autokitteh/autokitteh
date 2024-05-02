package dbfactory

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config = gormkitteh.Config

var Configs = configset.Set[Config]{
	Default: &Config{
		SlowQueryThreshold: 300 * time.Millisecond,
		Type:               gormkitteh.RequireExplicitDSNType,
	},
	Dev: &Config{
		DSN: "sqlite:file:" + filepath.Join(xdg.DataHomeDir(), "autokitteh.sqlite"),
	},
	Test: &Config{},
}

func New(z *zap.Logger, cfg *Config) (db.DB, error) {
	return dbgorm.New(z, cfg)
}

func NewTest(t *testing.T, objs []sdktypes.Object) db.DB {
	ctx := context.Background()

	z := zap.NewNop()

	testdb, err := New(z, &Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := testdb.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if err := testdb.Setup(ctx); err != nil {
		t.Fatalf("Setup: %v", err)
	}

	if err := db.Populate(ctx, testdb, objs...); err != nil {
		t.Fatalf("Populate: %v", err)
	}

	return testdb
}
