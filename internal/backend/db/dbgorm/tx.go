package dbgorm

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type txImpl struct {
	*gormdb
}

func (tx txImpl) LockProject(ctx context.Context, pid sdktypes.ProjectID) error {
	return translateError(
		tx.writer.
			Model(&scheme.Project{}).
			Where("id = ?", pid).
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
