package sdkclients

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkauthclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkbuildsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkconnectionsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkdeploymentsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkdispatcherclient"
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

		auth:         kittehs.LazyCache(sdkauthclient.New, params),
		builds:       kittehs.LazyCache(sdkbuildsclient.New, params),
		connections:  kittehs.LazyCache(sdkconnectionsclient.New, params),
		deployments:  kittehs.LazyCache(sdkdeploymentsclient.New, params),
		dispatcher:   kittehs.LazyCache(sdkdispatcherclient.New, params),
		events:       kittehs.LazyCache(sdkeventsclient.New, params),
		integrations: kittehs.LazyCache(sdkintegrationsclient.New, params),
		oauth:        kittehs.LazyCache(sdkoauthclient.New, params),
		projects:     kittehs.LazyCache(sdkprojectsclient.New, params),
		runtimes:     kittehs.LazyCache(sdkruntimesclient.New, params),
		sessions:     kittehs.LazyCache(sdksessionsclient.New, params),
		store:        kittehs.LazyCache(sdkstoreclient.New, params),
		triggers:     kittehs.LazyCache(sdktriggerclient.New, params),
		vars:         kittehs.LazyCache(sdkvarsclient.New, params),
	}
}

func (c *client) Auth() sdkservices.Auth                 { return c.auth() }
func (c *client) Builds() sdkservices.Builds             { return c.builds() }
func (c *client) Connections() sdkservices.Connections   { return c.connections() }
func (c *client) Deployments() sdkservices.Deployments   { return c.deployments() }
func (c *client) Dispatcher() sdkservices.Dispatcher     { return c.dispatcher() }
func (c *client) Events() sdkservices.Events             { return c.events() }
func (c *client) Integrations() sdkservices.Integrations { return c.integrations() }
func (c *client) OAuth() sdkservices.OAuth               { return c.oauth() }
func (c *client) Projects() sdkservices.Projects         { return c.projects() }
func (c *client) Runtimes() sdkservices.Runtimes         { return c.runtimes() }
func (c *client) Sessions() sdkservices.Sessions         { return c.sessions() }
func (c *client) Store() sdkservices.Store               { return c.store() }
func (c *client) Triggers() sdkservices.Triggers         { return c.triggers() }
func (c *client) Vars() sdkservices.Vars                 { return c.vars() }
