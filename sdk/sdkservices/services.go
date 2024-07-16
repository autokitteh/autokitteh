package sdkservices

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
