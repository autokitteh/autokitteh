package server

import (
	"context"

	"go.autokitteh.dev/autokitteh/backend/basesvc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

type db interface {
	Setup(context.Context) error
	Teardown(context.Context) error
}

func InitDB(mode string) (db, error) {
	return basesvc.StartDB(context.Background(), common.Config())
}
