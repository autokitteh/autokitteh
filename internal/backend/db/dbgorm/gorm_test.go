package dbgorm

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type dbFixture struct {
	db        *gorm.DB
	gormdb    *gormdb
	ctx       context.Context
	sessionID uint
}

func (f *dbFixture) newSessionID() (id uint) {
	id = f.sessionID
	f.sessionID += 1
	return id
}

// TODO: use gormkitteh (and maybe test with sqlite::memory and embedded PG)
func setupDB(dbName string) *gorm.DB {
	logger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,   // Slow SQL threshold
			LogLevel:      logger.Silent, // Log level
			Colorful:      false,         // Disable color
		},
	)
	if dbName == "" {
		dbName = ":memory:"
	}
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		NowFunc: func() time.Time { // operate always in UTC to simplify object comparison upon creation and fetching
			return time.Now().UTC()
		},
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	return db
}

func newDbFixture() *dbFixture {
	db := setupDB("") // in-memory db, specify filename to use file db

	ctx := context.Background()

	gormdb := gormdb{db: db, cfg: nil, mu: nil, z: zap.NewExample()}
	if err := gormdb.Teardown(ctx); err != nil { // delete tables if any
		log.Printf("Failed to termdown gormdb: %v", err)
	}
	if err := gormdb.Setup(ctx); err != nil { // ensure migration/schemas
		log.Fatalf("Failed to setup gormdb: %v", err)
	}

	return &dbFixture{db: db, gormdb: &gormdb, ctx: ctx}
}

// enable SQL logging
// func (f *dbFixture) debug() {
// 	f.db = f.db.Debug()
// 	f.gormdb.db = f.db
// }

func testDBExistsOne[T any](t *testing.T, gormdb *gormdb, schemaObj T, where string, args ...any) {
	var r []T
	res := gormdb.db.Where(where, args...).Find(&r)
	require.Nil(t, res.Error)
	require.Equal(t, 1, len(r))
	require.Equal(t, int64(1), res.RowsAffected)
	require.Equal(t, schemaObj, r[0])
}

var (
	testSessionID    = "ses_00000000000000000000000001"
	testDeploymentID = "dep_00000000000000000000000001"
	testEventID      = "evt_00000000000000000000000001"
	testBuildID      = "bld_00000000000000000000000001"
	testEnvID        = "env_00000000000000000000000001"
)

func newSession(f *dbFixture, st sdktypes.SessionStateType) scheme.Session {
	now := time.Now().UTC() // save and compare times in UTC
	sessionID := fmt.Sprintf("s:%04d", f.newSessionID())

	return scheme.Session{
		SessionID:        sessionID,
		DeploymentID:     testDeploymentID,
		EventID:          testEventID,
		CurrentStateType: int(st.ToProto()),
		Entrypoint:       "testEntrypoint",
		Inputs:           datatypes.JSON(`{"key": "value"}`),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func makeSchemeBuild() scheme.Build {
	now := time.Now().UTC() // save and compare times in UTC

	build := scheme.Build{
		BuildID:   testBuildID,
		Data:      []byte{},
		CreatedAt: now,
	}
	return build
}

func makeSchemeDeployment() scheme.Deployment {
	now := time.Now().UTC() // save and compare times in UTC

	deployment := scheme.Deployment{
		DeploymentID: testDeploymentID,
		BuildID:      testBuildID,
		EnvID:        testEnvID,
		State:        int32(sdktypes.DeploymentStateUnspecified.ToProto()),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return deployment
}
