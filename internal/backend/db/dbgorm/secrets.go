package dbgorm

import (
	"context"

	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (db *gormdb) SetSecret(ctx context.Context, key string, value string) error {
	secret := scheme.Secret{Key: key, Value: value}
	return translateError(db.wdb.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(secret).Error)
}

func (db *gormdb) GetSecret(ctx context.Context, key string) (string, error) {
	var secret scheme.Secret
	if err := db.rdb.WithContext(ctx).Where("key = ?", key).Find(&secret).Error; err != nil {
		return "", translateError(err)
	}

	return secret.Value, nil
}

func (db *gormdb) DeleteSecret(ctx context.Context, key string) error {
	return delete[scheme.Secret](db.wdb, ctx, "key = ?", key)
}
