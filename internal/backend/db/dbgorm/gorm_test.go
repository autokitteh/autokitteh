package dbgorm

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	embedPG "github.com/fergusstrange/embedded-postgres"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	now    time.Time
	dbType string
	gormDB gormdb
)

func TestMain(m *testing.M) {
	flag.StringVar(&dbType, "dbtype", "sqlite", "Database type to use for tests (e.g., sqlite, postgres)")
	flag.Parse()

	var pg *embedPG.EmbeddedPostgres

	if dbType == "postgres" {
		pg = embedPG.NewDatabase()
		if err := pg.Start(); err != nil {
			log.Fatalf("failed to start postgres: %v", err)
		}
		fmt.Println("Started PG..")
	}

	// setup test bench - gorm, schemas, migrations, etc
	cfg := gormkitteh.Config{Type: dbType, DSN: ""} // "" for in-memory, or specify a file
	db := setupDB(&cfg)
	gormDB = gormdb{db: db, cfg: &cfg, mu: nil, z: zap.NewExample()}

	ctx := context.Background()
	if err := gormDB.Setup(ctx); err != nil { // ensure migration/schemas
		log.Fatalf("Failed to setup gormdb: %v", err)
	}

	now = time.Now()
	now = now.Truncate(time.Microsecond) // PG default resolution is microseconds

	// PG saves dates in UTC. Gorm converts them back to local TZ on read
	// SQLite has no dedicated time format and uses strings, so gorm will read them as UTC, thus we need to write them in UTC
	if dbType == "sqlite" {
		now = now.UTC()
	}

	// run tests
	exitCode := m.Run()

	// teardown test bench
	if err := gormDB.Teardown(ctx); err != nil { // delete tables if any
		log.Printf("Failed to teardown gormdb: %v", err)
	}

	if dbType == "postgres" && pg != nil {
		// don't run this in defer to keep logs
		fmt.Println("Stopping PG...")
		if err := pg.Stop(); err != nil {
			log.Fatalf("failed to stop postgres: %v", err)
		}
	}

	os.Exit(exitCode)
}

func init() {
	now = time.Now()
	now = now.Truncate(time.Microsecond) // PG default resolution is microseconds
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
func setupDB(config *gormkitteh.Config) *gorm.DB {
	var dialector gorm.Dialector
	switch config.Type {
	case "sqlite":
		if config.DSN == "" {
			config.DSN = ":memory:"
		}
		dialector = sqlite.Open(config.DSN)
	case "postgres":
		dsn := "user=postgres password=postgres dbname=postgres host=localhost sslmode=disable"
		dialector = postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true, // disables implicit prepared statement usage. By default pgx automatically uses the extended protocol
		})
	default:
		log.Fatalf("unsuppported DBtype - <%s>", config.Type)
	}

	logger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,   // Slow SQL threshold
			LogLevel:      logger.Silent, // Log level
			Colorful:      false,         // Disable color
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{
		NowFunc: func() time.Time { // operate always in UTC to simplify object comparison upon creation and fetching
			return time.Now().UTC()
		},
		Logger:         logger,
		TranslateError: true,
		// DriverName:
	})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if config.Type == "sqlite" {
		db.Exec("PRAGMA foreign_keys = ON")
	}
	return db
}

func newDBFixture() *dbFixture {
	ctx := context.Background()
	if err := gormDB.Cleanup(ctx); err != nil { // ensure migration/schemas
		log.Fatalf("Failed to cleanup gormdb: %v", err)
	}
	return &dbFixture{db: gormDB.db, gormdb: &gormDB, ctx: ctx}
}

func newDBFixtureFK(withoutForeignKeys bool) *dbFixture {
	f := newDBFixture()
	if withoutForeignKeys { // run after setup, since this pragma may be reset by setup
		foreignKeys(f.gormdb, false)
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
		EventID:   eventID,
		CreatedAt: now,
	}
}

func (f *dbFixture) newEventRecord() scheme.EventRecord {
	eventID := fmt.Sprintf("evt_%026d", f.eventID)
	return scheme.EventRecord{
		EventID:   eventID,
		CreatedAt: now,
	}
}

func (f *dbFixture) newSignal() scheme.Signal {
	return scheme.Signal{
		SignalID:  testSignalID,
		CreatedAt: now,
	}
}
