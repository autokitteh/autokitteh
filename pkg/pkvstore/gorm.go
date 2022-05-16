package pkvstore

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

var _ Store = &GORMStore{}

type gormRecord struct {
	P string `gorm:"primaryKey"`
	K string `gorm:"primaryKey"`
	V []byte
}

func (g *GORMStore) db(ctx context.Context) *gorm.DB {
	return g.DB.WithContext(ctx).Table(g.TableName)
}

func (g *GORMStore) Put(ctx context.Context, p, k string, v []byte) error {
	if err := g.db(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&gormRecord{P: p, K: k, V: v}).Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (g *GORMStore) Get(ctx context.Context, p, k string) ([]byte, error) {
	var r gormRecord

	if err := g.db(ctx).First(&r, "k = ? AND p = ?", k, p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get: %w", err)
	}

	return r.V, nil
}

func (g *GORMStore) List(ctx context.Context, p string) ([]string, error) {
	var ks []string

	if err := g.db(ctx).Model(&gormRecord{}).Where("p = ?", p).Pluck("k", &ks).Error; err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	return ks, nil
}

func (g *GORMStore) Delete(ctx context.Context, p, k string) error {
	if err := g.db(ctx).Delete(&gormRecord{}, "k = ? AND p = ?", k, p).Error; err != nil {
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
