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
	Orgs() Orgs
	Projects() Projects
	Runtimes() Runtimes
	Secrets() Secrets
	Sessions() Sessions
	Store() Store
	Triggers() Triggers
	Users() Users
}
