package svc

import (
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type fxServices struct {
	fx.In

	Auth_         sdkservices.Auth         `optional:"true"`
	Builds_       sdkservices.Builds       `optional:"true"`
	Connections_  sdkservices.Connections  `optional:"true"`
	Deployments_  sdkservices.Deployments  `optional:"true"`
	Dispatcher_   sdkservices.Dispatcher   `optional:"true"`
	Envs_         sdkservices.Envs         `optional:"true"`
	Events_       sdkservices.Events       `optional:"true"`
	Integrations_ sdkservices.Integrations `optional:"true"`
	OAuth_        sdkservices.OAuth        `optional:"true"`
	Projects_     sdkservices.Projects     `optional:"true"`
	Runtimes_     sdkservices.Runtimes     `optional:"true"`
	Sessions_     sdkservices.Sessions     `optional:"true"`
	Store_        sdkservices.Store        `optional:"true"`
	Triggers_     sdkservices.Triggers     `optional:"true"`
	Vars_         sdkservices.Vars         `optional:"true"`
}

var _ sdkservices.Services = &fxServices{}

func (s *fxServices) Auth() sdkservices.Auth                 { return s.Auth_ }
func (s *fxServices) Builds() sdkservices.Builds             { return s.Builds_ }
func (s *fxServices) Connections() sdkservices.Connections   { return s.Connections_ }
func (s *fxServices) Deployments() sdkservices.Deployments   { return s.Deployments_ }
func (s *fxServices) Dispatcher() sdkservices.Dispatcher     { return s.Dispatcher_ }
func (s *fxServices) Envs() sdkservices.Envs                 { return s.Envs_ }
func (s *fxServices) Events() sdkservices.Events             { return s.Events_ }
func (s *fxServices) Integrations() sdkservices.Integrations { return s.Integrations_ }
func (s *fxServices) OAuth() sdkservices.OAuth               { return s.OAuth_ }
func (s *fxServices) Projects() sdkservices.Projects         { return s.Projects_ }
func (s *fxServices) Runtimes() sdkservices.Runtimes         { return s.Runtimes_ }
func (s *fxServices) Sessions() sdkservices.Sessions         { return s.Sessions_ }
func (s *fxServices) Store() sdkservices.Store               { return s.Store_ }
func (s *fxServices) Triggers() sdkservices.Triggers         { return s.Triggers_ }
func (s *fxServices) Vars() sdkservices.Vars                 { return s.Vars_ }
