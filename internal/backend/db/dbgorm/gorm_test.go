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
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var now time.Time

func assertErrorContainsNoCase(t *testing.T, err error, contains string) {
	assert.Contains(t, strings.ToUpper(err.Error()), strings.ToUpper(contains))
}

func init() {
	now = time.Now().UTC() // save and compare times in UTC
}

type dbFixture struct {
	db           *gorm.DB
	gormdb       *gormdb
	ctx          context.Context
	sessionID    uint
	deploymentID uint
	envID        uint
	projectID    uint
	triggerID    uint
	eventID      uint
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

func newDBFixture() *dbFixture {
	dsn := "file::memory:" // "/tmp/ak.db"
	db := setupDB(dsn)     // in-memory db, specify filename to use file db

	ctx := context.Background()
	cfg := gormkitteh.Config{Type: "sqlite", DSN: dsn}

	gormdb := gormdb{db: db, cfg: &cfg, mu: nil, z: zap.NewExample()}
	if err := gormdb.Teardown(ctx); err != nil { // delete tables if any
		log.Printf("Failed to termdown gormdb: %v", err)
	}
	if err := gormdb.Setup(ctx); err != nil { // ensure migration/schemas
		log.Fatalf("Failed to setup gormdb: %v", err)
	}
	return &dbFixture{db: db, gormdb: &gormdb, ctx: ctx}
}

func newDBFixtureFK(withoutForeignKeys bool) *dbFixture {
	f := newDBFixture()
	if withoutForeignKeys { // run after setup, since this pragma may be reset by setup
		f.db.Exec("PRAGMA foreign_keys = OFF")
	}
	return f
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
	require.ErrorIs(t, res.Error, gorm.ErrRecordNotFound)

	// check that object is marked as deleted
	res = f.db.Unscoped().First(&m)
	require.NoError(t, res.Error)
	require.Equal(t, int64(1), res.RowsAffected)

	deletedAtField := reflect.ValueOf(&m).Elem().FieldByName("DeletedAt")
	require.NotNil(t, deletedAtField.Interface())
}

// check obj is soft-deleted in gorm
func assertDeleted[T any](t *testing.T, f *dbFixture, m T) {
	// check that object is not found both scoped and unscoped
	res := f.db.First(&m)
	require.ErrorIs(t, res.Error, gorm.ErrRecordNotFound)

	// check that object is marked as deleted
	res = f.db.Unscoped().First(&m)
	require.ErrorIs(t, res.Error, gorm.ErrRecordNotFound)
}

var (
	// testSessionID    = "ses_00000000000000000000000001"
	testBuildID = "bld_00000000000000000000000001"
	// testDeploymentID = "dep_00000000000000000000000001"
	// testEventID      = "evt_00000000000000000000000001"
	testEnvID = "env_00000000000000000000000001"
	// testProjectID = "prj_00000000000000000000000001"
	// testTriggerID     = "trg_00000000000000000000000001"
	testConnectionID  = "con_00000000000000000000000001"
	testIntegrationID = "int_00000000000000000000000001"
	testSignalID      = "sig_00000000000000000000000001"
)

func (f *dbFixture) newSession(st sdktypes.SessionStateType) scheme.Session {
	f.sessionID += 1
	sessionID := fmt.Sprintf("ses_%026d", f.sessionID)

	return scheme.Session{
		SessionID:        sessionID,
		CurrentStateType: int(st.ToProto()),
		Inputs:           datatypes.JSON(`{"key": "value"}`),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (f *dbFixture) newSessionLogRecord() scheme.SessionLogRecord {
	sessionID := fmt.Sprintf("ses_%026d", f.sessionID)
	return scheme.SessionLogRecord{
		SessionID: sessionID,
	}
}

func (f *dbFixture) newBuild() scheme.Build {
	return scheme.Build{
		BuildID:   testBuildID,
		Data:      []byte{},
		CreatedAt: now,
	}
}

func (f *dbFixture) newDeployment() scheme.Deployment {
	f.deploymentID += 1
	deploymentID := fmt.Sprintf("dep_%026d", f.deploymentID)

	return scheme.Deployment{
		DeploymentID: deploymentID,
		State:        int32(sdktypes.DeploymentStateUnspecified.ToProto()),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (f *dbFixture) newProject() scheme.Project {
	f.projectID += 1
	projectID := fmt.Sprintf("prj_%026d", f.projectID)

	return scheme.Project{
		ProjectID: projectID,
		Name:      projectID, // must be unique
		Resources: []byte{},
	}
}

func (f *dbFixture) newEnv() scheme.Env {
	f.envID += 1
	envID := fmt.Sprintf("env_%026d", f.envID)

	return scheme.Env{
		EnvID:        envID,
		MembershipID: envID, // must be unique
	}
}

func (f *dbFixture) newTrigger() scheme.Trigger {
	f.triggerID += 1
	triggerID := fmt.Sprintf("trg_%026d", f.triggerID)

	return scheme.Trigger{
		TriggerID:    triggerID,
		EnvID:        testEnvID,
		ConnectionID: testConnectionID,
	}
}

func (f *dbFixture) newConnection() scheme.Connection {
	return scheme.Connection{
		ConnectionID: testConnectionID,
	}
}

func (f *dbFixture) newIntegration() scheme.Integration {
	return scheme.Integration{
		IntegrationID: testIntegrationID,
	}
}

func (f *dbFixture) newEvent() scheme.Event {
	f.eventID += 1
	eventID := fmt.Sprintf("evt_%026d", f.eventID)

	return scheme.Event{
		EventID: eventID,
	}
}

func (f *dbFixture) newEventRecord() scheme.EventRecord {
	eventID := fmt.Sprintf("evt_%026d", f.eventID)
	return scheme.EventRecord{
		EventID: eventID,
	}
}

func (f *dbFixture) newSignal() scheme.Signal {
	return scheme.Signal{
		SignalID: testSignalID,
	}
}
