package aksvc

import (
	"context"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/aksvc"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/svccommon"
)

type (
	RunOptions = aksvc.RunOptions
	Config     = svccommon.Config
)

var (
	LoadConfig = svccommon.LoadConfig
	StartDB    = aksvc.StartDB
)

type ShutdownSignal = fx.ShutdownSignal

type Service interface {
	Start(ctx context.Context) error
	Stop(context.Context) error
	Wait() <-chan ShutdownSignal
}

type service struct {
	app     *fx.App
	httpSvc httpsvc.Svc
}

func (s *service) Start(ctx context.Context) error { return s.app.Start(ctx) }
func (s *service) Stop(ctx context.Context) error  { return s.app.Stop(ctx) }
func (s *service) Wait() <-chan ShutdownSignal     { return s.app.Wait() }

func New(cfg *Config, ropts RunOptions) (Service, error) {
	var service service

	opts := append(
		aksvc.NewOpts(cfg, ropts),
		fx.Populate(&service.httpSvc),
	)

	service.app = fx.New(opts...)
	if err := service.app.Err(); err != nil {
		return nil, err
	}

	return &service, nil
}
