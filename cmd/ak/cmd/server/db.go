package server

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/svc"
)

type db interface {
	Setup(context.Context) error
	Migrate(context.Context) error
}

func InitDB(cfg *svc.Config, mode configset.Mode) (db, error) {
	return svc.StartDB(context.Background(), cfg, svc.RunOptions{Mode: mode})
}
