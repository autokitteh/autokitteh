package kvstore

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/pkg/gormfactory"
	L "github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/pkg/storefactory"
)

type Config = storefactory.Config

type Factory struct {
	Name string
}

func (f Factory) MustOpenString(ctx context.Context, l L.L, text string) Store {
	return f.MustOpen(ctx, l, storefactory.ParseConfigString(text))
}

func (f Factory) OpenString(ctx context.Context, l L.L, text string) (Store, error) {
	return f.Open(ctx, l, storefactory.ParseConfigString(text))
}

func (f Factory) MustOpen(ctx context.Context, l L.L, cfg *Config) Store {
	s, err := f.Open(ctx, l, cfg)
	if err != nil {
		panic(err)
	}
	return s
}

func (f Factory) Open(ctx context.Context, l L.L, cfg *Config) (Store, error) {
	r, err := mk(f.Name).Open(ctx, l, cfg)
	if err != nil {
		return nil, err
	}
	return r.(Store), nil
}

func mk(name string) *storefactory.Factory {
	inmem := func(ctx context.Context, _ L.L, _ *storefactory.InMemConfig) (interface{}, error) {
		db := gormfactory.MustOpenInMem()

		s := &GORMStore{DB: db, TableName: name}

		if err := s.Setup(ctx); err != nil {
			return nil, fmt.Errorf("setup: %w", err)
		}

		return s, nil
	}

	return &storefactory.Factory{
		FromDefault: func(ctx context.Context, l L.L) (interface{}, error) { return inmem(ctx, l, nil) },
		FromGORM: func(_ context.Context, _ L.L, db *gorm.DB) (interface{}, error) {
			return &GORMStore{DB: db, TableName: name}, nil
		},
		FromInMem: inmem,
		FromFS: func(_ context.Context, _ L.L, cfg *storefactory.FSConfig) (interface{}, error) {
			return &FSStore{RootPath: cfg.RootPath, Options: cfg.Options}, nil
		},
	}
}
