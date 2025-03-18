package server

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/aksvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/svccommon"
)

type db interface {
	Setup(context.Context) error
	Migrate(context.Context) error
}

func InitDB(cfg *svccommon.Config, mode configset.Mode) (db, error) {
	return aksvc.StartDB(context.Background(), cfg, aksvc.RunOptions{Mode: mode})
}
