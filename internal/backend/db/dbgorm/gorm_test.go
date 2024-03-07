package dbgorm

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
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

var now time.Time

func init() {
	now = time.Now().UTC() // save and compare times in UTC
}

type dbFixture struct {
	db     *gorm.DB
	gormdb *gormdb
	ctx    context.Context

	sessionID uint
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
	db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	return db
}

func newDbFixture(withoutForeignKeys bool) *dbFixture {
	db := setupDB("") // in-memory db, specify filename to use file db
	if withoutForeignKeys {
		db.Exec("PRAGMA foreign_keys = OFF")
	}

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

func findAndAssertCount[T any](t *testing.T, f *dbFixture, schemaObj T, expected int, where string, args ...any) []T {
	var objs []T
	res := f.gormdb.db.Where(where, args...).Find(&objs)
	require.NoError(t, res.Error)
	require.Equal(t, expected, len(objs))
	require.Equal(t, int64(expected), res.RowsAffected)
	return objs
}

func findAndAssertOne[T any](t *testing.T, f *dbFixture, schemaObj T, where string, args ...any) {
	res := findAndAssertCount(t, f, schemaObj, 1, where, args...)
	require.Equal(t, schemaObj, res[0])
}

// check obj is soft-deleted in gorm
func assertSoftDeleted[T any](t *testing.T, f *dbFixture, m T) {
	// check that object is not found without unscoped (due to deleted_at)
	res := f.db.First(&m)
	require.ErrorAs(t, gorm.ErrRecordNotFound, &res.Error)

	// check that object is marked as deleted
	res = f.db.Unscoped().First(&m)
	require.NoError(t, res.Error)
	require.Equal(t, int64(1), res.RowsAffected)
	res.Scan(&m)

	deletedAtField := reflect.ValueOf(&m).Elem().FieldByName("DeletedAt")
	require.NotNil(t, deletedAtField.Interface())
}

var (
	// testSessionID    = "ses_00000000000000000000000001"
	testBuildID      = "bld_00000000000000000000000001"
	testDeploymentID = "dep_00000000000000000000000001"
	testEventID      = "evt_00000000000000000000000001"
	testEnvID        = "env_00000000000000000000000001"
	testProjectID    = "prj_00000000000000000000000001"
	testProjectName  = "testProject"
)

func newSession(f *dbFixture, st sdktypes.SessionStateType) scheme.Session {
	f.sessionID += 1
	sessionID := fmt.Sprintf("ses_%026d", f.sessionID)

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

func newBuild() scheme.Build {
	return scheme.Build{
		BuildID:   testBuildID,
		Data:      []byte{},
		CreatedAt: now,
	}
}

func newDeployment(buildID string, envID string) scheme.Deployment {
	return scheme.Deployment{
		DeploymentID: testDeploymentID,
		BuildID:      buildID,
		EnvID:        envID,
		State:        int32(sdktypes.DeploymentStateUnspecified.ToProto()),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func newProject() scheme.Project {
	return scheme.Project{
		ProjectID: testProjectID,
		Name:      testProjectName,
		RootURL:   "",
		Resources: []byte{},
	}
}

func newEnv() scheme.Env {
	return scheme.Env{
		EnvID:        testEnvID,
		ProjectID:    testProjectID,
		Name:         "",
		MembershipID: "",
	}
}
