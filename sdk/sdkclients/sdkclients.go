package sdkclients

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkorgsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkprojectsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkruntimesclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdksecretsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdksessionsclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkstoreclient"
	sdktriggerclient "go.autokitteh.dev/autokitteh/sdk/sdkclients/sdktriggersclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkusersclient"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type client struct {
	params       sdkclient.Params
	auth         func() sdkservices.Auth
	builds       func() sdkservices.Builds
	connections  func() sdkservices.Connections
	deployments  func() sdkservices.Deployments
	dispatcher   func() sdkservices.Dispatcher
	envs         func() sdkservices.Envs
	events       func() sdkservices.Events
	integrations func() sdkservices.Integrations
	oauth        func() sdkservices.OAuth
	orgs         func() sdkservices.Orgs
	projects     func() sdkservices.Projects
	runtimes     func() sdkservices.Runtimes
	secrets      func() sdkservices.Secrets
	sessions     func() sdkservices.Sessions
	store        func() sdkservices.Store
	triggers     func() sdkservices.Triggers
	users        func() sdkservices.Users
}

func New(params sdkclient.Params) sdkservices.Services {
	return &client{
		auth:         kittehs.Lazy1(sdkauthclient.New, params),
		builds:       kittehs.Lazy1(sdkbuildsclient.New, params),
		connections:  kittehs.Lazy1(sdkconnectionsclient.New, params),
		deployments:  kittehs.Lazy1(sdkdeploymentsclient.New, params),
		dispatcher:   kittehs.Lazy1(sdkdispatcherclient.New, params),
		envs:         kittehs.Lazy1(sdkenvsclient.New, params),
		events:       kittehs.Lazy1(sdkeventsclient.New, params),
		integrations: kittehs.Lazy1(sdkintegrationsclient.New, params),
		oauth:        kittehs.Lazy1(sdkoauthclient.New, params),
		orgs:         kittehs.Lazy1(sdkorgsclient.New, params),
		params:       params, // just a dumb struct, no need to be lazy here.
		projects:     kittehs.Lazy1(sdkprojectsclient.New, params),
		runtimes:     kittehs.Lazy1(sdkruntimesclient.New, params),
		secrets:      kittehs.Lazy1(sdksecretsclient.New, params),
		sessions:     kittehs.Lazy1(sdksessionsclient.New, params),
		store:        kittehs.Lazy1(sdkstoreclient.New, params),
		triggers:     kittehs.Lazy1(sdktriggerclient.New, params),
		users:        kittehs.Lazy1(sdkusersclient.New, params),
	}
}

func ClientWithToken(c sdkservices.Services, t string) sdkservices.Services {
	v, ok := c.(*client)
	if !ok {
		sdklogger.Panic("original client is not a valid client")
	}

	params := v.params
	params.AuthToken = t

	return New(params)
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
func (c *client) Orgs() sdkservices.Orgs                 { return c.orgs() }
func (c *client) Projects() sdkservices.Projects         { return c.projects() }
func (c *client) Runtimes() sdkservices.Runtimes         { return c.runtimes() }
func (c *client) Secrets() sdkservices.Secrets           { return c.secrets() }
func (c *client) Sessions() sdkservices.Sessions         { return c.sessions() }
func (c *client) Store() sdkservices.Store               { return c.store() }
func (c *client) Triggers() sdkservices.Triggers         { return c.triggers() }
func (c *client) Users() sdkservices.Users               { return c.users() }
