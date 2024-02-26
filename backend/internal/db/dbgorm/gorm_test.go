package dbgorm

import (
	"context"
	"log"
	"os"
	"time"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type dbFixture struct {
	db     *gorm.DB
	gormdb *gormdb
	ctx    context.Context
}

// TODO: use akgorm layer instead of gorm directly
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

func getZ() *zap.Logger {
	// return zap.NewNop()
	logger := zap.NewExample()
	defer func() { // flushes buffer, if any
		if err := logger.Sync(); err != nil {
			log.Printf("Could not sync logger: %v", err)
		}
	}()
	return logger
}

func newDbFixture() *dbFixture {
	db := setupDB("") // in-memory db, specify filename to use file db

	ctx := context.Background()

	gormdb := gormdb{db: db, cfg: nil, mu: nil, z: getZ()}
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

var (
	testSessionID    = "s:1234"
	testDeploymentID = "d:1234"
	testEventID      = "ev:1234"
	testBuildID      = "b:1234"
	testEnvID        = "env:1234"
)

func makeSchemeSession() scheme.Session {
	now := time.Now().UTC() // save and compare times in UTC

	session := scheme.Session{
		SessionID:        testSessionID,
		DeploymentID:     testDeploymentID,
		EventID:          testEventID,
		CurrentStateType: int(sdktypes.CompletedSessionStateType),
		Entrypoint:       "testEntrypoint",
		Inputs:           datatypes.JSON(`{"key": "value"}`),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	return session
}

func makeSchemeDeployment() scheme.Deployment {
	now := time.Now().UTC() // save and compare times in UTC

	deployment := scheme.Deployment{
		DeploymentID: testDeploymentID,
		BuildID:      testBuildID,
		EnvID:        testEnvID,
		State:        int32(sdktypes.DeploymentStateUnspecified),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return deployment
}
