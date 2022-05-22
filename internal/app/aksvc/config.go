package aksvc

import (
	"github.com/autokitteh/autokitteh/internal/app/dashboardsvc"
	"github.com/autokitteh/autokitteh/internal/app/githubeventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/app/googleoauthsvc"
	"github.com/autokitteh/autokitteh/internal/app/temporalite"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/akprocs"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore/eventsrcsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore/eventsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/githubinstalls"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/projectsstorefactory"
	"github.com/autokitteh/autokitteh/internal/pkg/sessions"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore/statestorefactory"

	"github.com/autokitteh/autokitteh/internal/app/croneventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/app/slackeventsrcsvc"
	"github.com/autokitteh/autokitteh/internal/pkg/fseventsrc"
	"github.com/autokitteh/autokitteh/internal/pkg/httpeventsrc"

	"github.com/autokitteh/autokitteh/pkg/initd"

	"github.com/autokitteh/L"
	"github.com/autokitteh/stores/kvstore"
	"github.com/autokitteh/stores/pkvstore"
	"github.com/autokitteh/stores/storefactory"
)

type TemporalConfig struct {
	HostPort  string `envconfig:"HOSTPORT" json:"hostport" default:"localhost:7233"`
	Namespace string `envconfig:"NAMESPACE" json:"namespace" default:"default"`
}

type Config struct {
	Initd             initd.Config                 `envconfig:"INITD" json:"initd"`
	EmbeddedDash      bool                         `envconfig:"EMBEDDED_DASH" default:"true" json:"embedded_dash"`
	InitPaths         []string                     `envconfig:"INIT_PATHS" json:"init_paths"`
	Temporal          TemporalConfig               `envconfig:"TEMPORAL" json:"temporal"`
	Temporalite       temporalite.Config           `envconfig:"TEMPORALITE" json:"temporalite"`
	CatalogPermissive bool                         `envconfig:"CATALOG_PERMISSIVE" json:"catalog_permissive"`
	DefaultStore      storefactory.Config          `envconfig:"DEFAULT_STORE" json:"default_store"`
	AccountsStore     accountsstorefactory.Config  `envconfig:"ACCOUNTS_STORE" json:"accounts_store"`
	ProjectsStore     projectsstorefactory.Config  `envconfig:"PROJECTS_STORE" json:"projects_store"`
	StateStore        statestorefactory.Config     `envconfig:"STATE_STORE" json:"state_store"`
	EventsStore       eventsstorefactory.Config    `envconfig:"EVENTS_STORE" json:"events_store"`
	EventSourcesStore eventsrcsstorefactory.Config `envconfig:"EVENT_SOURCES_STORE" json:"event_srcs_store"`
	Sessions          sessions.Config              `envconfig:"SESSIONS" json:"sessions"`
	Dashboard         dashboardsvc.Config          `envconfig:"DASHBOARD" json:"dashboard"`
	UtilityStore      kvstore.Config               `envconfig:"UTILITY_STORE" json:"utility_store"`
	SecretsStore      pkvstore.Config              `envconfig:"SECRETS_STORE" json:"secrets_store"`
	CredsStore        pkvstore.Config              `envconfig:"CREDS_STORE" json:"creds_store"`
	PluginsRegStore   pkvstore.Config              `envconfig:"PLUGINS_REG_STORE" json:"plugins_reg_store"`
	PluginsRegProcs   akprocs.Config               `envconfig:"PLUGINS_REG_PROCS" json:"plugins_reg_procs"`

	// [# google-oauth-config #]
	GoogleOAuthSvc            googleoauthsvc.Config `envconfig:"GOOGLE_OAUTH" json:"googleauths"`
	GoogleOAuthSvcTokensStore kvstore.Config        `envconfig:"GOOGLE_OAUTH_TOKENS_STORE" json:"googleauths_svc_oauth_tokens_store"`

	CronEventSource                  croneventsrcsvc.Config   `envconfig:"CRON_EVENT_SOURCE" json:"cron_event_src"`
	DefaultTwilioAccountSid          string                   `envconfig:"DEFAULT_TWILIO_ACCOUNT_SID" json:"default_twilio_account_sid"`
	DefaultTwilioAuthToken           string                   `envconfig:"DEFAULT_TWILIO_AUTH_TOKEN" json:"default_twilio_auth_token"`
	HTTPEventSource                  httpeventsrc.Config      `envconfig:"HTTP_EVENT_SOURCE" json:"http_event_src"`
	FSEventSource                    fseventsrc.Config        `envconfig:"FS_EVENT_SOURCE" json:"fs_event_src"`
	SlackEventSource                 slackeventsrcsvc.Config  `envconfig:"SLACK_EVENT_SOURCE" json:"slack_event_src"`
	SlackEventSourceOAuthTokensStore kvstore.Config           `envconfig:"SLACK_EVENT_SOURCE_OAUTH_TOKENS_STORE" json:"slack_oauth_tokens_store"`
	DefaultSlackTeamIDs              map[string]string        `envconfig:"DEFAULT_SLACK_TEAM_IDS" json:"default_slack_team_ids"`
	GithubEventSource                githubeventsrcsvc.Config `envconfig:"GITHUB_EVENT_SOURCE" json:"github_event_src"`
	GithubInstallsStore              kvstore.Config           `envconfig:"GITHUB_INSTALLS_STORE" json:"github_installs_store"`
	GithubInstalls                   githubinstalls.Config    `envconfig:"GITHUB_INSTALLS" json:"github_installs"`
	DefaultGithubRepos               []string                 `envconfig:"DEFAULT_GITHUB_REPOS" json:"default_github_repos"`
}

func (c *Config) PostSvcLoad(l L.L) error {
	def := func(what string, s *storefactory.Config) {
		if !s.IsSet() && c.DefaultStore.IsSet() {
			*s = c.DefaultStore

			l.Named(what).Debug("using default store config")
		}
	}

	def("accountsstore", &c.AccountsStore)
	def("projectsstore", &c.ProjectsStore)
	def("statestore", &c.StateStore)
	def("eventsstore", &c.EventsStore)
	def("eventsrcsstore", &c.EventSourcesStore)
	def("eventsrcsstore", &c.UtilityStore)
	def("slackoauthtokensstore", &c.SlackEventSourceOAuthTokensStore)
	def("googleoauthtokensstore", &c.GoogleOAuthSvcTokensStore)
	def("utilitystore", &c.UtilityStore)
	def("secretsstore", &c.SecretsStore)
	def("credsstore", &c.CredsStore)
	def("pluginsregstore", &c.PluginsRegStore)
	def("githubinstalls", &c.GithubInstallsStore)

	// Set from [# InitPaths #] in cli.go
	c.InitPaths = append(c.InitPaths, initPaths.Value()...)

	if c.PluginsRegProcs.ReadyAddress == "" {
		c.PluginsRegProcs.ReadyAddress = "http://127.0.0.1:20000/pluginsreg/ready"
	}

	return nil
}
