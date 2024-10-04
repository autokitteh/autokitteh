package server

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
	"go.autokitteh.dev/autokitteh/internal/backend/svc"
)

type db interface {
	Setup(context.Context) error
	Migrate(context.Context) error
}

func InitDB(cfg *config.Config, mode config.Mode) (db, error) {
	return svc.StartDB(context.Background(), cfg, svc.RunOptions{Mode: mode})
}
