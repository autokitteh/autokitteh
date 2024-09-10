package server

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/config"
)

type db interface {
	Setup(context.Context) error
	Migrate(context.Context) error
}

func InitDB(cfg *config.Config, mode configset.Mode) (db, error) {
	return svc.StartDB(context.Background(), cfg, svc.RunOptions{Mode: mode})
}
