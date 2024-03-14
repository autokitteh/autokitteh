package dbgorm

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/db"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
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
	default:
		return fmt.Errorf("db: %w", err)
	}
}

func (db *gormdb) Setup(ctx context.Context) error {
	if err := db.db.WithContext(ctx).AutoMigrate(scheme.Tables...); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	for _, s := range seeds {
		if err := s(ctx, db); err != nil {
			return err
		}
	}

	return nil
}

func (db *gormdb) Teardown(ctx context.Context) error {
	if err := db.db.WithContext(ctx).Migrator().DropTable(scheme.Tables...); err != nil {
		return fmt.Errorf("droptable: %w", err)
	}

	return nil
}

// Todo: not sure this will work with the connect method
func (db *gormdb) Debug() db.DB {
	return &gormdb{
		z:  db.z,
		db: db.db.Debug(),
	}
}

func get[T, R any](db *gorm.DB, ctx context.Context, f func(t T) (R, error), where string, args ...any) (R, error) {
	var (
		t T
		r R
	)

	result := db.WithContext(ctx).Where(where, args...).Limit(1).Find(&t)
	if result.Error != nil {
		return r, translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return r, sdkerrors.ErrNotFound
	}

	return f(t)
}
