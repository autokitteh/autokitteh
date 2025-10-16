package dbgorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/migrator"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type Config = gormkitteh.Config

type gormdb struct {
	z   *zap.Logger
	cfg *Config

	writer, reader *gorm.DB
}

// Migrate implements db.DB.
func (db *gormdb) Migrate(ctx context.Context) error {
	_, c, err := db.client(false)
	if err != nil {
		return err
	}
	m, err := migrator.NewMigrator(c, db.cfg.Type)
	if err != nil {
		return err
	}

	return m.Migrate(ctx)
}

var _ db.DB = (*gormdb)(nil)

func New(z *zap.Logger, cfg *Config) (db.DB, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg, err := cfg.Explicit()
	if err != nil {
		return nil, err
	}

	return &gormdb{z: z, cfg: cfg}, nil
}

func (db *gormdb) GormDB() (r, w *gorm.DB) { return db.reader, db.writer }

func connect(_ context.Context, z *zap.Logger, cfg *Config) (r *gorm.DB, w *gorm.DB, err error) {
	gormCfgFn := func(cfg *gorm.Config) { cfg.SkipDefaultTransaction = true }

	r, err = gormkitteh.OpenZ(z, cfg, gormCfgFn)
	if err != nil {
		err = fmt.Errorf("opendb: %w", err)
		return
	}

	var sqlDB *sql.DB
	if sqlDB, err = r.DB(); err != nil {
		return
	}

	n := cfg.MaxOpenConns
	if n == 0 {
		if cfg.InferredType() == "sqlite" {
			n = 1
		} else {
			n = min(4, runtime.NumCPU())
		}
	}

	sqlDB.SetMaxOpenConns(n)

	// For in memory sqlite in memory database we will use the same connection for reads and writes
	// since otherwise it will have two distinct databases for these.
	if cfg.Type == "sqlite" && strings.HasPrefix(cfg.DSN, ":memory:") {
		w = r
		return
	}

	// For SQlite we need to open a separate connection for writes.
	// See https://kerkour.com/sqlite-for-servers.

	w, err = gormkitteh.OpenZ(z, cfg, gormCfgFn)
	if err != nil {
		err = fmt.Errorf("opendb: %w", err)
		return
	}

	if sqlDB, err = w.DB(); err != nil {
		return
	}

	sqlDB.SetMaxOpenConns(1)

	return
}

func (db *gormdb) Connect(ctx context.Context) (err error) {
	db.reader, db.writer, err = connect(ctx, db.z.Named("gorm"), db.cfg)
	return
}

func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows):
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

func gormErrNotFoundToForeignKey(err error) error {
	if err == gorm.ErrRecordNotFound {
		return gorm.ErrForeignKeyViolated
	}
	return err
}

var fkStmtByDB = map[string]map[bool]string{
	"sqlite": {
		true:  "PRAGMA foreign_keys = ON",
		false: "PRAGMA foreign_keys = OFF",
	},
	"postgres": {
		// in PG foreign keys implemented as triggers. Setting `session_replication_role'
		// to `replica' prevents firing triggers, thus effectively disables foreign keys
		true:  "SET session_replication_role = DEFAULT",
		false: "SET session_replication_role = replica",
	},
}

func foreignKeys(gormdb *gormdb, enable bool) error {
	if _, found := fkStmtByDB[gormdb.cfg.Type]; !found {
		panic(fmt.Errorf("unknown DB type: %s", gormdb.cfg.Type))
	}
	stmt := fkStmtByDB[gormdb.cfg.Type][enable]

	if err := gormdb.reader.Exec(stmt).Error; err != nil {
		return fmt.Errorf("read exec: %q: %w", stmt, err)
	}

	if gormdb.writer != gormdb.reader {
		if err := gormdb.writer.Exec(stmt).Error; err != nil {
			return fmt.Errorf("write exec: %q: %w", stmt, err)
		}
	}

	return nil
}

func (db *gormdb) seed(ctx context.Context) error {
	if db.cfg.SeedCommands == "" {
		return nil
	}

	db.z.Info("seeding")

	cmd := db.writer.WithContext(ctx).Debug().Exec(db.cfg.SeedCommands)

	db.z.Info("done seeding", zap.Int64("rows_affected", cmd.RowsAffected))

	return translateError(cmd.Error)
}

func (db *gormdb) Setup(ctx context.Context) error {
	isSqlite := db.cfg.Type == "sqlite"
	if isSqlite {
		if err := foreignKeys(db, false); err != nil {
			return err
		}

		defer func() {
			if err := foreignKeys(db, true); err != nil {
				db.z.Error("failed to re-enable foreign keys", zap.Error(err))
			}
		}()
	}
	_, c, err := db.client(true)
	if err != nil {
		return err
	}

	m, err := migrator.NewMigrator(c, db.cfg.Type)
	if err != nil {
		return err
	}

	requireMigration, err := m.MigrationRequired(ctx)
	if err != nil {
		return err
	}

	if requireMigration || m.DbVersion == 0 {
		if err := m.Migrate(ctx); err != nil {
			return err
		}
	}

	if err := db.seed(ctx); err != nil {
		return err
	}

	return nil
}

// TODO: not sure this will work with the connect method
func (db *gormdb) Debug() db.DB {
	return &gormdb{
		z:      db.z,
		reader: db.reader.Debug(),
		writer: db.writer.Debug(),
	}
}

// NOTE: no ctx is passed since in all places it's already applied
func getOne[T any](db *gorm.DB, where string, args ...any) (*T, error) {
	var r []T
	result := db.Where(where, args...).Limit(2).Find(&r)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if result.RowsAffected > 1 {
		return nil, gorm.ErrDuplicatedKey
	}
	return &r[0], nil
}

// TODO: this not working for deployments. Consider delete this function
func delete[T any](db *gorm.DB, ctx context.Context, where string, args ...any) error {
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

func (db *gormdb) client(debug bool) (r, w *sql.DB, err error) {
	q := db.reader
	if debug {
		q = q.Debug()
	}

	if r, err = q.DB(); err != nil {
		return
	}

	q = db.writer
	if debug {
		q = q.Debug()
	}

	w, err = q.DB()

	return
}
