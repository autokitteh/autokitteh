package dbgorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "ariga.io/atlas-provider-gorm/gormschema"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/migrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config = gormkitteh.Config

type gormdb struct {
	z   *zap.Logger
	db  *gorm.DB
	cfg *Config

	// See https://github.com/mattn/go-sqlite3/issues/274.
	// Used only for protecting writes when using sqlite.
	//
	// TODO(ENG-190): This is not... ideal. Really find a better
	//                solution. Everything else I tried didn't work.
	mu *sync.Mutex
}

var _ db.DB = (*gormdb)(nil)

func New(z *zap.Logger, cfg *Config) (db.DB, error) {
	cfg, err := cfg.Explicit()
	if err != nil {
		return nil, err
	}

	db := &gormdb{z: z, cfg: cfg}

	if cfg.Type == "sqlite" {
		db.mu = new(sync.Mutex)
	}

	return db, nil
}

func (db *gormdb) GormDB() *gorm.DB { return db.db }

func (db *gormdb) Connect(ctx context.Context) error {
	client, err := gormkitteh.OpenZ(db.z.Named("gorm"), db.cfg, func(cfg *gorm.Config) {
		cfg.SkipDefaultTransaction = true
		cfg.Logger = logger.Default
	})
	if err != nil {
		return fmt.Errorf("opendb: %w", err)
	}

	db.db = client
	return nil
}

func (db *gormdb) locked(f func(db *gormdb) error) error {
	if db.mu != nil {
		db.mu.Lock()
		defer db.mu.Unlock()
	}

	return translateError(f(db))
}

func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return sdkerrors.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return sdkerrors.ErrAlreadyExists
	case errors.Is(err, sdkerrors.ErrAlreadyExists):
		return err
	case errors.Is(err, sdkerrors.ErrNotFound):
		return err
	default:
		return fmt.Errorf("db: %w", err)
	}
}

func foreignKeys(gormdb *gormdb, enable bool) {
	var stmt string
	var valMap map[bool]string
	if gormdb.cfg.Type == "sqlite" {
		stmt = "PRAGMA foreign_keys = %s"
		valMap = map[bool]string{true: "ON", false: "OFF"}
	} else if gormdb.cfg.Type == "postgres" {
		stmt = "SET session_replication_role = %s;"
		valMap = map[bool]string{true: "replica", false: "DEFAULT"}
	}
	gormdb.db.Exec(fmt.Sprintf(stmt, valMap[enable]))

}

func initGoose(client *sql.DB, dialect string) error {
	goose.SetBaseFS(migrations.Migrations)

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if _, err := goose.EnsureDBVersion(client); err != nil {
		return fmt.Errorf("failed to ensure DB version: %w", err)
	}
	return nil
}

func (db *gormdb) Migrate(ctx context.Context) error {

	client := db.client()

	if err := initGoose(client, db.cfg.Type); err != nil {
		return err
	}

	migrationsDir := db.cfg.Type
	return goose.Up(client, migrationsDir)
}

func (db *gormdb) MigrationRequired(ctx context.Context) (bool, int64, error) {
	client := db.client()
	if err := initGoose(client, db.cfg.Type); err != nil {
		return false, 0, err
	}

	dbversion, err := goose.GetDBVersion(client)
	if err != nil {
		return false, 0, err
	}

	migrationsDir := db.cfg.Type
	requiredMigrations, err := goose.CollectMigrations(migrationsDir, dbversion, int64((1<<63)-1))
	if err != nil && !errors.Is(err, goose.ErrNoMigrationFiles) {
		return false, 0, err
	}

	return len(requiredMigrations) > 0, dbversion, nil
}

func (db *gormdb) Setup(ctx context.Context) error {
	isSqlite := db.cfg.Type == "sqlite"
	if isSqlite {
		db.db.Exec("PRAGMA foreign_keys = OFF")
		defer func() {
			db.db.Exec("PRAGMA foreign_keys = ON")
		}()
	}

	required, dbVersion, err := db.MigrationRequired(ctx)
	if err != nil {
		return err
	}
	if !required {
		return nil
	}

	return errors.New("db migrations required") //TODO: maybe more details
}

func (db *gormdb) Teardown(ctx context.Context) error {
	isSqlite := db.cfg.Type == "sqlite"
	if isSqlite {
		foreignKeys(db, false)
	}
	if err := db.db.WithContext(ctx).Migrator().DropTable(scheme.Tables...); err != nil {
		return fmt.Errorf("droptable: %w", err)
	}
	if isSqlite {
		foreignKeys(db, true)
	}

	return nil
}

// TODO: not sure this will work with the connect method
func (db *gormdb) Debug() db.DB {
	return &gormdb{
		z:  db.z,
		db: db.db.Debug(),
	}
}

func getOneWTransform[T any, R sdktypes.Object](db *gorm.DB, ctx context.Context, f func(t T) (R, error), where string, args ...any) (R, error) {
	var (
		rec     T
		invalid R
	)

	// TODO: fetch all records and report if there is more than one record
	result := db.WithContext(ctx).Where(where, args...).Limit(1).Find(&rec)
	if result.Error != nil {
		return invalid, translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return invalid, sdkerrors.ErrNotFound
	}

	return f(rec)
}

// TODO: change all get functions to use this
func getOne[T any](db *gorm.DB, ctx context.Context, t T, where string, args ...any) (*T, error) {
	var r T

	// TODO: fetch all records and report if there is more than one record
	result := db.WithContext(ctx).Where(where, args...).Limit(1).Find(&r)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &r, nil
}

// TODO: this not working for deployments. Consider delete this function
func delete[T any](db *gorm.DB, ctx context.Context, t T, where string, args ...any) error {
	var r T
	result := db.WithContext(ctx).Where(where, args...).Delete(&r)
	if result.Error != nil {
		return translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return sdkerrors.ErrNotFound
	}

	return nil
}

func (db *gormdb) client() *sql.DB {
	return kittehs.Must1(db.db.DB())
}
