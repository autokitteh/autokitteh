package dbgorm

import (
	"context"
	"encoding/json"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func (db *gormdb) SetSecret(ctx context.Context, key string, value map[string]string) error {
	secret := scheme.Secret{Key: key, Value: kittehs.Must1(json.Marshal(value))}
	return translateError(db.db.WithContext(ctx).Create(secret).Error)
}

func (db *gormdb) GetSecret(ctx context.Context, key string) (map[string]string, error) {
	var secret scheme.Secret
	if err := db.db.WithContext(ctx).Where("key = ?", key).Find(&secret).Error; err != nil {
		return nil, translateError(err)
	}

	var data map[string]string
	if err := json.Unmarshal(secret.Value, &data); err != nil {
		return nil, translateError(err)
	}

	return data, nil

}

func (db *gormdb) DeleteSecret(ctx context.Context, key string) error {
	return delete(db.db, ctx, scheme.Secret{}, "key = ?", key)
}
