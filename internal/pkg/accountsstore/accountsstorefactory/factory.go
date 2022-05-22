package accountsstorefactory

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	pbaccountsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/accountsvc"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstoregorm"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstoregrpc"
	"github.com/autokitteh/stores/gormfactory"
	"github.com/autokitteh/stores/storefactory"
)

type Config = storefactory.Config

func newInMem(ctx context.Context, _ L.L) (accountsstore.Store, error) {
	db := gormfactory.MustOpenInMem()

	as := &accountsstoregorm.Store{DB: db}

	if err := as.Setup(ctx); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}

	return as, nil
}

var f = storefactory.Factory{
	FromDefault: func(ctx context.Context, l L.L) (interface{}, error) {
		return newInMem(ctx, l)
	},
	FromInMem: func(ctx context.Context, l L.L, _ *storefactory.InMemConfig) (interface{}, error) {
		return newInMem(ctx, l)
	},
	FromGORM: func(ctx context.Context, _ L.L, db *gorm.DB) (interface{}, error) {
		return &accountsstoregorm.Store{DB: db}, nil
	},
	FromGRPCConn: func(_ context.Context, _ L.L, conn *grpc.ClientConn) (interface{}, error) {
		return &accountsstoregrpc.Store{Client: pbaccountsvc.NewAccountsClient(conn)}, nil
	},
}

func OpenString(ctx context.Context, l L.L, text string) (accountsstore.Store, error) {
	store, err := f.OpenString(ctx, l, text)
	if err != nil {
		return nil, err
	}

	return store.(accountsstore.Store), nil
}

func Open(ctx context.Context, l L.L, cfg *Config) (accountsstore.Store, error) {
	store, err := f.Open(ctx, l, cfg)
	if err != nil {
		return nil, err
	}

	return store.(accountsstore.Store), nil
}
