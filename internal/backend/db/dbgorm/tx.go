package dbgorm

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
)

type tx struct {
	gormdb
	done bool
}

func (db *gormdb) Begin(ctx context.Context) (db.Transaction, error) {
	db.db.Error = nil
	tx1 := db.db.WithContext(ctx).Begin()
	if err := tx1.Error; err != nil {
		return nil, err
	}

	return &tx{
		gormdb: gormdb{
			z:  db.z.With(zap.String("txid", uuid.NewString())),
			db: tx1,
		},
	}, nil
}

func (tx *tx) Commit() error {
	tx.done = true
	return tx.db.Commit().Error
}

func (tx *tx) Rollback() error {
	if tx.done {
		return nil
	}

	return tx.db.Rollback().Error
}

func (db *gormdb) Transaction(ctx context.Context, f func(db db.DB) error) error {
	return db.transaction(ctx, func(tx *tx) error {
		return f(tx)
	})
}

func (db *gormdb) transaction(ctx context.Context, f func(tx *tx) error) error {
	return db.locked(func(db *gormdb) error {
		return db.db.WithContext(ctx).Transaction(func(txdb *gorm.DB) error {
			return f(
				&tx{
					gormdb: gormdb{
						z:  db.z.With(zap.String("txid", uuid.NewString())),
						db: txdb,
					},
				},
			)
		})
	})
}

// FIXME: fix/rewrite locked()
// the same as transaction, but without translating the error in locked()
func (db *gormdb) transaction2(ctx context.Context, f func(tx *tx) error) error {
	return db.locked2(func(db *gormdb) error {
		return db.db.WithContext(ctx).Transaction(func(txdb *gorm.DB) error {
			return f(
				&tx{
					gormdb: gormdb{
						z:  db.z.With(zap.String("txid", uuid.NewString())),
						db: txdb,
					},
				},
			)
		})
	})
}
