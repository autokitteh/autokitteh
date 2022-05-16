package statestorefactory

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore/statestoregorm"
	"github.com/autokitteh/autokitteh/pkg/gormfactory"
	L "github.com/autokitteh/autokitteh/pkg/l"
	"github.com/autokitteh/autokitteh/pkg/storefactory"
)

type Config = storefactory.Config

func newInMem(ctx context.Context, _ L.L) (interface{}, error) {
	db := gormfactory.MustOpenInMem()

	es := &statestoregorm.Store{DB: db}

	if err := es.Setup(ctx); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}

	return es, nil
}

var factory = storefactory.Factory{
	FromDefault: func(ctx context.Context, l L.L) (interface{}, error) {
		return newInMem(ctx, l)
	},
	FromInMem: func(ctx context.Context, l L.L, _ *storefactory.InMemConfig) (interface{}, error) {
		return newInMem(ctx, l)
	},
	FromGORM: func(ctx context.Context, _ L.L, db *gorm.DB) (interface{}, error) {
		return &statestoregorm.Store{DB: db}, nil
	},
	// TODO: redis.
}

func OpenString(ctx context.Context, l L.L, text string) (statestore.Store, error) {
	store, err := factory.OpenString(ctx, l, text)
	if err != nil {
		return nil, err
	}

	return store.(statestore.Store), nil
}

func Open(ctx context.Context, l L.L, cfg *Config) (statestore.Store, error) {
	store, err := factory.Open(ctx, l, cfg)
	if err != nil {
		return nil, err
	}

	return store.(statestore.Store), nil
}
