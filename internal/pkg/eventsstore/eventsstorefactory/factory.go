package eventsstorefactory

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	pbeventsvc "go.autokitteh.dev/idl/go/eventsvc"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore/eventsstoregorm"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore/eventsstoregrpc"
	"github.com/autokitteh/stores/gormfactory"
	"github.com/autokitteh/stores/storefactory"
)

type Config = storefactory.Config

func newInMem(ctx context.Context, _ L.L) (interface{}, error) {
	db := gormfactory.MustOpenInMem()

	es := &eventsstoregorm.Store{DB: db}

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
		return &eventsstoregorm.Store{DB: db}, nil
	},
	FromGRPCConn: func(ctx context.Context, _ L.L, conn *grpc.ClientConn) (interface{}, error) {
		return &eventsstoregrpc.Store{Client: pbeventsvc.NewEventsClient(conn)}, nil
	},
}

func OpenString(ctx context.Context, l L.L, text string) (eventsstore.Store, error) {
	store, err := factory.OpenString(ctx, l, text)
	if err != nil {
		return nil, err
	}

	return store.(eventsstore.Store), nil
}

func Open(ctx context.Context, l L.L, cfg *Config) (eventsstore.Store, error) {
	store, err := factory.Open(ctx, l, cfg)
	if err != nil {
		return nil, err
	}

	return store.(eventsstore.Store), nil
}
