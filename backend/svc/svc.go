package svc

import (
	"context"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/basesvc"
	"go.autokitteh.dev/autokitteh/internal/backend/svc"
)

type (
	RunOptions = basesvc.RunOptions
	Config     = basesvc.Config
)

var (
	LoadConfig = basesvc.LoadConfig
	StartDB    = basesvc.StartDB
)

type ShutdownSignal = fx.ShutdownSignal

type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Wait() <-chan ShutdownSignal
}

type service struct{ *fx.App }

func (s *service) Start(ctx context.Context) error { return s.App.Start(ctx) }
func (s *service) Stop(ctx context.Context) error  { return s.App.Stop(ctx) }
func (s *service) Wait() <-chan ShutdownSignal     { return s.App.Wait() }

func New(cfg *Config, ropts RunOptions) (Service, error) {
	app := fx.New(svc.NewOpts(cfg, ropts)...)
	if app.Err() != nil {
		return nil, app.Err()
	}

	return &service{app}, nil
}
