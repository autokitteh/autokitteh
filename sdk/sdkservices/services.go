package sdkservices

import "go.uber.org/fx"

type Services interface {
	DBServices

	Auth() Auth
	Dispatcher() Dispatcher
	OAuth() OAuth
	Runtimes() Runtimes
	Store() Store
}

type DBServices interface {
	Integrations() Integrations
	Projects() Projects
	Builds() Builds
	Deployments() Deployments
	Envs() Envs
	Connections() Connections
	Sessions() Sessions
	Events() Events
	Triggers() Triggers
	Vars() Vars
}

type ServicesStruct struct {
	fx.In // this can also be used using uber's Fx.

	Auth_         Auth         `optional:"true"`
	Builds_       Builds       `optional:"true"`
	Connections_  Connections  `optional:"true"`
	Deployments_  Deployments  `optional:"true"`
	Dispatcher_   Dispatcher   `optional:"true"`
	Envs_         Envs         `optional:"true"`
	Events_       Events       `optional:"true"`
	Integrations_ Integrations `optional:"true"`
	OAuth_        OAuth        `optional:"true"`
	Projects_     Projects     `optional:"true"`
	Runtimes_     Runtimes     `optional:"true"`
	Sessions_     Sessions     `optional:"true"`
	Store_        Store        `optional:"true"`
	Triggers_     Triggers     `optional:"true"`
	Vars_         Vars         `optional:"true"`
}

var _ Services = &ServicesStruct{}

func (s *ServicesStruct) Auth() Auth                 { return s.Auth_ }
func (s *ServicesStruct) Builds() Builds             { return s.Builds_ }
func (s *ServicesStruct) Connections() Connections   { return s.Connections_ }
func (s *ServicesStruct) Deployments() Deployments   { return s.Deployments_ }
func (s *ServicesStruct) Dispatcher() Dispatcher     { return s.Dispatcher_ }
func (s *ServicesStruct) Envs() Envs                 { return s.Envs_ }
func (s *ServicesStruct) Events() Events             { return s.Events_ }
func (s *ServicesStruct) Integrations() Integrations { return s.Integrations_ }
func (s *ServicesStruct) OAuth() OAuth               { return s.OAuth_ }
func (s *ServicesStruct) Projects() Projects         { return s.Projects_ }
func (s *ServicesStruct) Runtimes() Runtimes         { return s.Runtimes_ }
func (s *ServicesStruct) Sessions() Sessions         { return s.Sessions_ }
func (s *ServicesStruct) Store() Store               { return s.Store_ }
func (s *ServicesStruct) Triggers() Triggers         { return s.Triggers_ }
func (s *ServicesStruct) Vars() Vars                 { return s.Vars_ }
