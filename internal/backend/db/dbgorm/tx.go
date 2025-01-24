package dbgorm

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
)

func (db *gormdb) Transaction(ctx context.Context, f func(db db.DB) error) error {
	return db.transaction(ctx, func(tx *gormdb) error { return f(tx) })
}

func (db *gormdb) transaction(ctx context.Context, f func(tx *gormdb) error) error {
	return db.wdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return f(
			&gormdb{
				z:   db.z.With(zap.String("txid", uuid.NewString())),
				wdb: tx,
				rdb: tx,
				cfg: db.cfg,
			},
		)
	})
}

func (db *gormdb) rtransaction(ctx context.Context, f func(tx *gormdb) error) error {
	return db.rdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return f(
			&gormdb{
				z:   db.z.With(zap.String("txid", uuid.NewString())),
				wdb: nil, // panic on writes.
				rdb: tx,
				cfg: db.cfg,
			},
		)
	})
}
