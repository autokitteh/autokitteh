package dbgorm

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

type txImpl struct {
	*gormdb
}

func (db *gormdb) PrepareLock(ctx context.Context, id string) error {
	return db.writer.WithContext(ctx).Create(scheme.Lock{ID: id}).Error
}

func (tx txImpl) Lock(ctx context.Context, id string) error {
	return translateError(
		tx.writer.
			Model(&scheme.Lock{}).
			Where("id = ?", id).
			UpdateColumn("count", gorm.Expr("count + ?", 1)).
			Error,
	)
}

func (db *gormdb) Transaction(ctx context.Context, f func(db db.TX) error) error {
	return db.writeTransaction(ctx, func(tx *gormdb) error { return f(txImpl{tx}) })
}

func (db *gormdb) writeTransaction(ctx context.Context, f func(tx *gormdb) error) error {
	return db.writer.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return f(
			&gormdb{
				z:      db.z.With(zap.String("txid", uuid.NewString())),
				writer: tx,
				reader: tx,
				cfg:    db.cfg,
			},
		)
	})
}

/* readTransaction is a helper function to run a read-only transaction.
   for now unused, but keeping it here so we can use it in the future.

func (db *gormdb) readTransaction(ctx context.Context, f func(tx *gormdb) error) error {
	return db.reader.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return f(
			&gormdb{
				z:      db.z.With(zap.String("txid", uuid.NewString())),
				writer: nil, // panic on writes.
				reader: tx,
				cfg:    db.cfg,
			},
		)
	})
}
*/
