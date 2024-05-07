package sdkservices

type Services interface {
	Auth() Auth
	Builds() Builds
	Connections() Connections
	Deployments() Deployments
	Dispatcher() Dispatcher
	Envs() Envs
	Events() Events
	Integrations() Integrations
	OAuth() OAuth
	Projects() Projects
	Runtimes() Runtimes
	Sessions() Sessions
	Store() Store
	Triggers() Triggers
	Vars() Vars
}
