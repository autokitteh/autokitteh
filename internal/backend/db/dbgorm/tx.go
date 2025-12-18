package dbgorm

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
)

func (db *gormdb) Transaction(ctx context.Context, f func(db db.DB) error) error {
	return db.writeTransaction(ctx, func(tx *gormdb) error { return f(tx) })
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
