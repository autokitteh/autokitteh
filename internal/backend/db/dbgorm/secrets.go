package dbgorm

import (
	"context"
	"encoding/json"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (db *gormdb) SetSecret(ctx context.Context, name string, data map[string]string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	s := scheme.Secret{Name: name, Data: jsonData}
	result := db.db.WithContext(ctx).Create(&s)
	if result.Error != nil {
		return translateError(result.Error)
	}
	return nil
}

func (db *gormdb) GetSecret(ctx context.Context, name string) (map[string]string, error) {
	// Why Take() and not First()? We don't care about order because at
	// most a single record exists (we retrieve based on the primary key).
	// Also, if the primary key isn't found, Take() returns an error
	s := scheme.Secret{}
	result := db.db.WithContext(ctx).Take(&s, "name = ?", name)
	if result.Error != nil {
		return nil, translateError(result.Error)
	}

	data := make(map[string]string)
	if err := json.Unmarshal(s.Data, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (db *gormdb) AppendSecret(ctx context.Context, name, token string) error {
	// No need to combine both queries into a single transaction,
	// because only the last one affects the data.
	data, err := db.GetSecret(ctx, name)
	if err != nil {
		return err
	}

	data[token] = time.Now().UTC().Format(time.RFC3339)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	result := db.db.WithContext(ctx).Model(&scheme.Secret{}).Where("name = ?", name).Update("data", jsonData)
	if result.Error != nil {
		return translateError(result.Error)
	}
	return nil
}

func (db *gormdb) DeleteSecret(ctx context.Context, name string) error {
	// Reminder: Delete() is idempotent, i.e. no error if PK not found.
	result := db.db.WithContext(ctx).Delete(&scheme.Secret{}, "name = ?", name)
	if result.Error != nil {
		return translateError(result.Error)
	}
	return nil
}
