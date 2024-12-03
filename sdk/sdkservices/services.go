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
	Builds() Builds
	Connections() Connections
	Deployments() Deployments
	Events() Events
	Integrations() Integrations
	Orgs() Orgs
	Projects() Projects
	Sessions() Sessions
	Triggers() Triggers
	Users() Users
	Vars() Vars
}

type ServicesStruct struct {
	fx.In // this can also be used using uber's Fx.

	Auth_         Auth         `optional:"true"`
	Builds_       Builds       `optional:"true"`
	Connections_  Connections  `optional:"true"`
	Deployments_  Deployments  `optional:"true"`
	Dispatcher_   Dispatcher   `optional:"true"`
	Events_       Events       `optional:"true"`
	Integrations_ Integrations `optional:"true"`
	OAuth_        OAuth        `optional:"true"`
	Orgs_         Orgs         `optional:"true"`
	Projects_     Projects     `optional:"true"`
	Runtimes_     Runtimes     `optional:"true"`
	Sessions_     Sessions     `optional:"true"`
	Store_        Store        `optional:"true"`
	Triggers_     Triggers     `optional:"true"`
	Users_        Users        `optional:"true"`
	Vars_         Vars         `optional:"true"`
}

var _ Services = &ServicesStruct{}

func (s *ServicesStruct) Auth() Auth                 { return s.Auth_ }
func (s *ServicesStruct) Builds() Builds             { return s.Builds_ }
func (s *ServicesStruct) Connections() Connections   { return s.Connections_ }
func (s *ServicesStruct) Deployments() Deployments   { return s.Deployments_ }
func (s *ServicesStruct) Dispatcher() Dispatcher     { return s.Dispatcher_ }
func (s *ServicesStruct) Events() Events             { return s.Events_ }
func (s *ServicesStruct) Integrations() Integrations { return s.Integrations_ }
func (s *ServicesStruct) OAuth() OAuth               { return s.OAuth_ }
func (s *ServicesStruct) Orgs() Orgs                 { return s.Orgs_ }
func (s *ServicesStruct) Projects() Projects         { return s.Projects_ }
func (s *ServicesStruct) Runtimes() Runtimes         { return s.Runtimes_ }
func (s *ServicesStruct) Sessions() Sessions         { return s.Sessions_ }
func (s *ServicesStruct) Store() Store               { return s.Store_ }
func (s *ServicesStruct) Triggers() Triggers         { return s.Triggers_ }
func (s *ServicesStruct) Users() Users               { return s.Users_ }
func (s *ServicesStruct) Vars() Vars                 { return s.Vars_ }
