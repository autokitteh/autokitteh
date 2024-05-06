package sdkservices

type Services interface {
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
	Secrets() Secrets
	Sessions() Sessions
	Store() Store
	Triggers() Triggers
	Vars() Vars
}
