package server

import (
	"context"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type db interface {
	Setup(context.Context) error
	Teardown(context.Context) error
}

func InitDB(mode string) (db, error) {
	return svc.StartDB(context.Background(), common.Config(), svc.RunOptions{Mode: configset.Mode(mode)})
}
