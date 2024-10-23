package sdkclients

import (
	"sync"

	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkauthclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkbuildsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkconnectionsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkdeploymentsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkdispatcherclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkenvsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkeventsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkintegrationsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkoauthclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkprojectsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkruntimesclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdksessionsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkstoreclient"
	sdktriggerclient "go.autokitteh.dev/autokitteh/sdk/sdkclients/sdktriggersclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkvarsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type client struct {
	auth         func() sdkservices.Auth
	builds       func() sdkservices.Builds
	connections  func() sdkservices.Connections
	deployments  func() sdkservices.Deployments
	dispatcher   func() sdkservices.Dispatcher
	envs         func() sdkservices.Envs
	events       func() sdkservices.Events
	integrations func() sdkservices.Integrations
	oauth        func() sdkservices.OAuth
	params       sdkclient.Params
	projects     func() sdkservices.Projects
	runtimes     func() sdkservices.Runtimes
	sessions     func() sdkservices.Sessions
	store        func() sdkservices.Store
	triggers     func() sdkservices.Triggers
	vars         func() sdkservices.Vars
}

func New(params sdkclient.Params) sdkservices.Services {
	return &client{
		params: params, // just a dumb struct, no need to be lazy here.

		auth:         lazyCache(sdkauthclient.New, params),
		builds:       lazyCache(sdkbuildsclient.New, params),
		connections:  lazyCache(sdkconnectionsclient.New, params),
		deployments:  lazyCache(sdkdeploymentsclient.New, params),
		dispatcher:   lazyCache(sdkdispatcherclient.New, params),
		envs:         lazyCache(sdkenvsclient.New, params),
		events:       lazyCache(sdkeventsclient.New, params),
		integrations: lazyCache(sdkintegrationsclient.New, params),
		oauth:        lazyCache(sdkoauthclient.New, params),
		projects:     lazyCache(sdkprojectsclient.New, params),
		runtimes:     lazyCache(sdkruntimesclient.New, params),
		sessions:     lazyCache(sdksessionsclient.New, params),
		store:        lazyCache(sdkstoreclient.New, params),
		triggers:     lazyCache(sdktriggerclient.New, params),
		vars:         lazyCache(sdkvarsclient.New, params),
	}
}

func (c *client) Auth() sdkservices.Auth                 { return c.auth() }
func (c *client) Builds() sdkservices.Builds             { return c.builds() }
func (c *client) Connections() sdkservices.Connections   { return c.connections() }
func (c *client) Deployments() sdkservices.Deployments   { return c.deployments() }
func (c *client) Dispatcher() sdkservices.Dispatcher     { return c.dispatcher() }
func (c *client) Envs() sdkservices.Envs                 { return c.envs() }
func (c *client) Events() sdkservices.Events             { return c.events() }
func (c *client) Integrations() sdkservices.Integrations { return c.integrations() }
func (c *client) OAuth() sdkservices.OAuth               { return c.oauth() }
func (c *client) Projects() sdkservices.Projects         { return c.projects() }
func (c *client) Runtimes() sdkservices.Runtimes         { return c.runtimes() }
func (c *client) Sessions() sdkservices.Sessions         { return c.sessions() }
func (c *client) Store() sdkservices.Store               { return c.store() }
func (c *client) Triggers() sdkservices.Triggers         { return c.triggers() }
func (c *client) Vars() sdkservices.Vars                 { return c.vars() }

// lazyCache wraps a function and a single input. The first call to the wrapper
// calls the wrapped function. Subsequent calls return the first result.
func lazyCache[T, P any](f func(P) T, p P) func() T {
	var (
		t    T
		once sync.Once
	)

	return func() T {
		once.Do(func() { t = f(p) })

		return t
	}
}
