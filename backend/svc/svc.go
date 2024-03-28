package svc

import (
	"context"
	"net/http"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/svc"
)

type (
	RunOptions = svc.RunOptions
	Config     = svc.Config
)

var (
	LoadConfig = svc.LoadConfig
	StartDB    = svc.StartDB
)

type ShutdownSignal = fx.ShutdownSignal

type Service interface {
	Start(context.Context) error
	Stop(context.Context) error
	Wait() <-chan ShutdownSignal
	Addr() string // available after Start.
}

type service struct {
	app  *fx.App
	mux  *http.ServeMux
	addr svc.HTTPServerAddr
}

func (s *service) Start(ctx context.Context) error { return s.app.Start(ctx) }
func (s *service) Stop(ctx context.Context) error  { return s.app.Stop(ctx) }
func (s *service) Wait() <-chan ShutdownSignal     { return s.app.Wait() }
func (s *service) Addr() string                    { return string(s.addr) }

func New(cfg *Config, ropts RunOptions) (Service, error) {
	var service service

	opts := append(
		svc.NewOpts(cfg, ropts),
		fx.Populate(&service.mux),
		fx.Populate(&service.addr),
	)

	service.app = fx.New(opts...)
	if service.app.Err() != nil {
		return nil, service.app.Err()
	}

	return &service, nil
}
