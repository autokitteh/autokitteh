package server

import (
	"context"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

type db interface {
	Setup(context.Context) error
	Migrate(context.Context) error
}

func InitDB(mode string) (db, error) {
	return svc.StartDB(context.Background(), common.Config())
}
