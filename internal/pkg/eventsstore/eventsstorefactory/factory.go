package eventsstorefactory

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	pbeventsvc "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/eventsvc"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsstore"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsstore/eventsstoregorm"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsstore/eventsstoregrpc"
	"gitlab.com/softkitteh/autokitteh/pkg/gormfactory"
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
	"gitlab.com/softkitteh/autokitteh/pkg/storefactory"
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
