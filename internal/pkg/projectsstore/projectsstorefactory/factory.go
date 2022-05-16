package projectsstorefactory

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	pbprojectsvc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/projectsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/projectsstoregorm"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/projectsstoregrpc"
	"github.com/autokitteh/autokitteh/pkg/gormfactory"
	L "github.com/autokitteh/autokitteh/pkg/l"
	"github.com/autokitteh/autokitteh/pkg/storefactory"
)

type Config = storefactory.Config

func newInMem(ctx context.Context, _ L.L, accounts accountsstore.Store) (projectsstore.Store, error) {
	db := gormfactory.MustOpenInMem()

	ps := &projectsstoregorm.Store{DB: db, AccountsStore: accounts}

	if err := ps.Setup(ctx); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}

	return ps, nil
}

func factory(accountsStore accountsstore.Store) *storefactory.Factory {
	return &storefactory.Factory{
		FromDefault: func(ctx context.Context, l L.L) (interface{}, error) {
			return newInMem(ctx, l, accountsStore)
		},
		FromInMem: func(ctx context.Context, l L.L, _ *storefactory.InMemConfig) (interface{}, error) {
			return newInMem(ctx, l, accountsStore)
		},
		FromGORM: func(ctx context.Context, _ L.L, db *gorm.DB) (interface{}, error) {
			return &projectsstoregorm.Store{DB: db, AccountsStore: accountsStore}, nil
		},
		FromGRPCConn: func(_ context.Context, _ L.L, conn *grpc.ClientConn) (interface{}, error) {
			return &projectsstoregrpc.Store{Client: pbprojectsvc.NewProjectsClient(conn)}, nil
		},
	}
}

func OpenString(ctx context.Context, l L.L, text string, accountsStore accountsstore.Store) (projectsstore.Store, error) {
	out, err := factory(accountsStore).OpenString(ctx, l, text)
	if err != nil {
		return nil, err
	}

	return out.(projectsstore.Store), nil
}

func Open(ctx context.Context, l L.L, cfg *Config, accountsStore accountsstore.Store) (projectsstore.Store, error) {
	out, err := factory(accountsStore).Open(ctx, l, cfg)
	if err != nil {
		return nil, err
	}

	return out.(projectsstore.Store), nil
}
