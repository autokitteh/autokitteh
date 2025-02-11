package dbgorm

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	embedPG "github.com/fergusstrange/embedded-postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	now    time.Time
	dbType string
	gormDB gormdb
	id     int64 // running testID

	testIntegrationID = sdktypes.NewIntegrationIDFromName("test").UUIDValue()
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
	cfg := &gormkitteh.Config{Type: dbType, DSN: ""} // "" for in-memory, or specify a file
	cfg, _ = cfg.Explicit()
	db := setupDB(cfg)
	z := kittehs.Must1(zap.NewDevelopment())
	gormDB = gormdb{db: db, cfg: cfg, mu: nil, z: z}

	ctx := context.Background()
	if err := gormDB.Setup(ctx); err != nil { // ensure migration/schemas
		log.Fatalf("Failed to setup gormdb: %v", err)
	}

	now = kittehs.Now()
	now = now.Truncate(time.Microsecond) // PG default resolution is microseconds

	// PG saves dates in UTC. Gorm converts them back to local TZ on read
	// SQLite has no dedicated time format and uses strings, so gorm will read them as UTC, thus we need to write them in UTC
	if dbType == "sqlite" {
		now = now.UTC()
	}

	// run tests
	exitCode := m.Run()

	// teardown test bench
	if err := TeardownDB(&gormDB, ctx); err != nil { // delete tables if any
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
	now = kittehs.Now()
	now = now.Truncate(time.Microsecond) // PG default resolution is microseconds
}

type dbFixture struct {
	db            *gorm.DB
	gormdb        *gormdb
	ctx           context.Context
	eventSequence int
}

func newTestID() uuid.UUID {
	bytes := make([]byte, 16)
	id += 1
	binary.BigEndian.PutUint64(bytes[8:], uint64(id)) // fill the last 8 bytes, leave the first 8 bytes as zero

	return kittehs.Must1(uuid.FromBytes(bytes))
}

func idToName(id uuid.UUID, prefix string) string {
	idStr := id.String()
	if len(idStr) > 10 {
		idStr = idStr[len(idStr)-10:]
	}
	return fmt.Sprintf("%s_%s", prefix, idStr)
}

// TODO: use gormkitteh (and maybe test with sqlite::memory and embedded PG)
func setupDB(config *gormkitteh.Config) *gorm.DB {
	// mimic gormkitteh.Open
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
		log.Fatalf("unsupported DBtype - <%s>", config.Type)
	}

	logger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel:      logger.Silent,          // Log level
			Colorful:      false,                  // Disable color
			SlowThreshold: time.Millisecond * 500, // Slow SQL threshold
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

func TeardownDB(gormdb *gormdb, ctx context.Context) error {
	isSqlite := gormdb.cfg.Type == "sqlite"
	if isSqlite {
		foreignKeys(gormdb, false)
	}
	db := gormdb.db.WithContext(ctx).Scopes(NoDebug())
	if err := db.Migrator().DropTable(scheme.Tables...); err != nil {
		return fmt.Errorf("droptable: %w", err)
	}
	if err := db.Migrator().DropTable("goose_db_version"); err != nil {
		return fmt.Errorf("failed to drop migraiton table (goose_db_version) : %w", err)
	}
	if isSqlite {
		foreignKeys(gormdb, true)
	}

	return nil
}

func CleanupDB(gormdb *gormdb, ctx context.Context) error {
	foreignKeys(gormdb, false) // disable foreign keys

	db := gormdb.db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true})
	for _, model := range scheme.Tables {
		modelType := reflect.TypeOf(model)
		model := reflect.New(modelType).Interface()
		if err := db.Delete(model).Error; err != nil {
			if !strings.Contains(err.Error(), "no such table") {
				return fmt.Errorf("cleanup data for table %s: %w", modelType.Name(), err)
			}
		}
	}

	foreignKeys(gormdb, true) // re-enable foreign keys
	return nil
}

func newDBFixture() *dbFixture {
	ctx := context.Background()
	if err := CleanupDB(&gormDB, ctx); err != nil { // ensure migration/schemas
		log.Fatalf("Failed to cleanup gormdb: %v", err)
	}
	gormdb := gormDB
	f := dbFixture{db: gormdb.db, gormdb: &gormdb, ctx: ctx}
	return &f
}

func (f *dbFixture) WithForeignKeysDisabled(fn func()) {
	foreignKeys(f.gormdb, false) // disable
	fn()
	foreignKeys(f.gormdb, true) // enable
}

// enable SQL logging
func (f *dbFixture) WithDebug() *dbFixture {
	f.db = f.db.Debug()
	f.gormdb.db = f.db
	return f
}

func NoDebug() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		config := db.Config
		config.Logger = logger.Default.LogMode(logger.Warn)
		db.Config = config
		return db
	}
}

func findAndAssertCount[T any](t *testing.T, f *dbFixture, expected int, where string, args ...any) []T {
	var objs []T
	res := f.gormdb.db.Where(where, args...).Find(&objs)
	require.NoError(t, res.Error)
	require.Equal(t, expected, len(objs))
	require.Equal(t, int64(expected), res.RowsAffected)
	return objs
}

func findAndAssertOne[T any](t *testing.T, f *dbFixture, schemaObj T, where string, args ...any) {
	res := findAndAssertCount[T](t, f, 1, where, args...)
	resetTimes(&res[0], &schemaObj)
	require.Equal(t, schemaObj, res[0])
}

// check obj is soft-deleted in gorm
func assertSoftDeleted[T any](t *testing.T, f *dbFixture, m T) {
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
	testDummyID  = kittehs.Must1(uuid.FromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	testSignalID = kittehs.Must1(uuid.NewV7())
)

func (f *dbFixture) newSession(args ...any) scheme.Session {
	s := scheme.Session{
		SessionID: newTestID(),
		// CurrentStateType: int(st.ToProto()),
		Inputs:     datatypes.JSON([]byte("{}")),
		Entrypoint: "loc",
		Base: scheme.Base{
			CreatedAt: now,
		},
	}
	for _, a := range args {
		switch a := a.(type) {
		case sdktypes.SessionStateType:
			s.CurrentStateType = int(a.ToProto())
		case scheme.Build:
			s.BuildID = a.BuildID
		case scheme.Deployment:
			s.DeploymentID = &a.DeploymentID
		case scheme.Project:
			s.ProjectID = a.ProjectID
		case scheme.Event:
			s.EventID = &a.EventID
		}
	}

	resetTimes(&s)

	return s
}

func (f *dbFixture) newSessionLogRecord(sid uuid.UUID) scheme.SessionLogRecord {
	return scheme.SessionLogRecord{
		SessionID: sid,
	}
}

func (f *dbFixture) newBuild(args ...any) scheme.Build {
	b := scheme.Build{
		BuildID: newTestID(),
		Data:    []byte{},
		Base:    scheme.Base{CreatedAt: now},
	}
	for _, a := range args {
		if a, ok := a.(scheme.Project); ok {
			b.ProjectID = a.ProjectID
		}
	}
	return b
}

func (f *dbFixture) newDeployment(args ...any) scheme.Deployment {
	d := scheme.Deployment{
		DeploymentID: newTestID(),
		State:        int32(sdktypes.DeploymentStateUnspecified.ToProto()),
		Base:         scheme.Base{CreatedAt: now},
	}
	for _, a := range args {
		switch a := a.(type) {
		case scheme.Build:
			d.BuildID = a.BuildID
		case scheme.Project:
			d.ProjectID = a.ProjectID
		}
	}

	resetTimes(&d)
	return d
}

func (f *dbFixture) newProject() scheme.Project {
	id := newTestID()
	return scheme.Project{
		ProjectID: id,
		Name:      idToName(id, "prj"),
		Resources: []byte{},
	}
}

func (f *dbFixture) newVar(name string, val string, args ...any) scheme.Var {
	v := scheme.Var{
		ScopeID: testDummyID,
		Name:    name,
		Value:   val,
	}
	for _, a := range args {
		switch a := a.(type) {
		case scheme.Connection:
			v.ScopeID = a.ConnectionID
			v.IntegrationID = *a.IntegrationID
		case scheme.Project:
			v.ScopeID = a.ProjectID
		case uuid.UUID:
			v.ScopeID = a
		}
	}
	v.VarID = v.ScopeID
	return v
}

func (f *dbFixture) newTrigger(args ...any) scheme.Trigger {
	id := newTestID()
	name := idToName(id, "trg")
	t := scheme.Trigger{
		TriggerID:    id,
		Name:         name,
		CodeLocation: "loc",
		SourceType:   sdktypes.TriggerSourceTypeConnection.String(),
	}
	for _, a := range args {
		switch a := a.(type) {
		case scheme.Project:
			t.ProjectID = a.ProjectID
		case scheme.Connection:
			t.ConnectionID = &a.ConnectionID
		case string:
			t.Name = a
		}
	}
	t.UniqueName = triggerUniqueName(t.ProjectID.String(), sdktypes.NewSymbol(t.Name))
	return t
}

func (f *dbFixture) newConnection(args ...any) scheme.Connection {
	id := newTestID()
	name := idToName(id, "con")
	c := scheme.Connection{
		IntegrationID: &testIntegrationID,
		ConnectionID:  id,
		Name:          name,
	}
	for _, a := range args {
		switch a := a.(type) {
		case scheme.Project:
			c.ProjectID = a.ProjectID
		case string:
			c.Name = a
		}
	}
	return c
}

func (f *dbFixture) newEvent(args ...any) scheme.Event {
	f.eventSequence++
	e := scheme.Event{
		EventID: newTestID(),
		Base:    scheme.Base{CreatedAt: now},
		Seq:     uint64(f.eventSequence),
		Data:    kittehs.Must1(json.Marshal(struct{}{})),
		Memo:    kittehs.Must1(json.Marshal(struct{}{})),
	}
	for _, a := range args {
		switch a := a.(type) {
		case scheme.Connection:
			e.ConnectionID = &a.ConnectionID
		case scheme.Trigger:
			e.TriggerID = &a.TriggerID
		case scheme.Project:
			e.ProjectID = a.ProjectID
		}
	}
	return e
}

func (f *dbFixture) newSignal(args ...any) scheme.Signal {
	s := scheme.Signal{
		SignalID:  testSignalID,
		CreatedAt: now,
	}
	for _, a := range args {
		if a, ok := a.(scheme.Connection); ok {
			s.ConnectionID = &a.ConnectionID
			s.DestinationID = a.ConnectionID
		}
	}
	return s
}

// Reset all time.Time fields to zero. This is needed to compare objects with time fields.
// Each v in vs must be a pointer to a struct.
func resetTimes(vs ...any) {
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			panic("expected a pointer")
		}

		rv = rv.Elem()

		for i := range rv.NumField() {
			fv := rv.Field(i)
			if fv.Kind() == reflect.Struct {
				if fv.Type().Name() == "Time" {
					fv.Set(reflect.ValueOf(time.Time{}))
				} else {
					resetTimes(fv.Addr().Interface())
				}
			}
		}
	}
}

func TestResetTimes(t *testing.T) {
	type X struct {
		Time time.Time
	}
	type T struct {
		Time time.Time
		X    X
	}
	t1 := T{Time: time.Now(), X: X{Time: time.Now()}}
	resetTimes(&t1)
	require.Equal(t, time.Time{}, t1.Time)
	require.Equal(t, time.Time{}, t1.X.Time)
}
