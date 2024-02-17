package basesvc

import (
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type fxServices struct {
	fx.In

	Builds_       sdkservices.Builds       `optional:"true"`
	Connections_  sdkservices.Connections  `optional:"true"`
	Deployments_  sdkservices.Deployments  `optional:"true"`
	Dispatcher_   sdkservices.Dispatcher   `optional:"true"`
	Envs_         sdkservices.Envs         `optional:"true"`
	Events_       sdkservices.Events       `optional:"true"`
	Integrations_ sdkservices.Integrations `optional:"true"`
	Mappings_     sdkservices.Mappings     `optional:"true"`
	OAuth_        sdkservices.OAuth        `optional:"true"`
	Projects_     sdkservices.Projects     `optional:"true"`
	Runtimes_     sdkservices.Runtimes     `optional:"true"`
	Secrets_      sdkservices.Secrets      `optional:"true"`
	Sessions_     sdkservices.Sessions     `optional:"true"`
	Store_        sdkservices.Store        `optional:"true"`
}

var _ sdkservices.Services = &fxServices{}

func (s *fxServices) Builds() sdkservices.Builds             { return s.Builds_ }
func (s *fxServices) Connections() sdkservices.Connections   { return s.Connections_ }
func (s *fxServices) Deployments() sdkservices.Deployments   { return s.Deployments_ }
func (s *fxServices) Dispatcher() sdkservices.Dispatcher     { return s.Dispatcher_ }
func (s *fxServices) Envs() sdkservices.Envs                 { return s.Envs_ }
func (s *fxServices) Events() sdkservices.Events             { return s.Events_ }
func (s *fxServices) Integrations() sdkservices.Integrations { return s.Integrations_ }
func (s *fxServices) Mappings() sdkservices.Mappings         { return s.Mappings_ }
func (s *fxServices) OAuth() sdkservices.OAuth               { return s.OAuth_ }
func (s *fxServices) Projects() sdkservices.Projects         { return s.Projects_ }
func (s *fxServices) Runtimes() sdkservices.Runtimes         { return s.Runtimes_ }
func (s *fxServices) Secrets() sdkservices.Secrets           { return s.Secrets_ }
func (s *fxServices) Sessions() sdkservices.Sessions         { return s.Sessions_ }
func (s *fxServices) Store() sdkservices.Store               { return s.Store_ }
