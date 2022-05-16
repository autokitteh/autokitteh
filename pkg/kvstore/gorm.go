package kvstore

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMStore struct {
	DB        *gorm.DB
	TableName string
}

type gormRecord struct {
	Key   string `gorm:"primaryKey"`
	Value []byte
}

func (g *GORMStore) db(ctx context.Context) *gorm.DB {
	return g.DB.WithContext(ctx).Table(g.TableName)
}

func (g *GORMStore) Put(ctx context.Context, k string, v []byte) error {
	if err := g.db(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&gormRecord{Key: k, Value: v}).Error; err != nil {
		return fmt.Errorf("create error: %w", err)
	}

	return nil
}

func (g *GORMStore) Get(ctx context.Context, key string) ([]byte, error) {
	var r gormRecord

	if err := g.db(ctx).First(&r, "key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get error: %w", err)
	}

	return r.Value, nil
}

func (g *GORMStore) Delete(ctx context.Context, key string) error {
	if err := g.db(ctx).Delete(&gormRecord{}, "key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}

		return fmt.Errorf("get error: %w", err)
	}

	return nil
}

func (g *GORMStore) Setup(ctx context.Context) error {
	if err := g.db(ctx).AutoMigrate(&gormRecord{}); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	return nil
}

func (g *GORMStore) Teardown(ctx context.Context) error {
	if err := g.db(ctx).Migrator().DropTable(&gormRecord{}); err != nil {
		return fmt.Errorf("drop: %w", err)
	}

	return nil
}
