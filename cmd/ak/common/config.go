package common

import (
	"net/url"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
)

const (
	EnvVarPrefix       = "AK_"
	ConfigYAMLFileName = "config.yaml"
)

var (
	cfg       *svc.Config
	serverURL *url.URL
)

func ConfigYAMLFilePath() string {
	return filepath.Join(xdg.ConfigHomeDir(), ConfigYAMLFileName)
}

func InitConfig(confmap map[string]any) (err error) {
	// Reminder: cfg is a package-scoped variable, not function-scoped.
	if cfg, err = svc.LoadConfig(EnvVarPrefix, confmap, ConfigYAMLFilePath()); err != nil {
		return
	}

	serverURL, err = readServerURL()

	return
}

func Config() *svc.Config { return cfg }

func ServerURL() *url.URL { return serverURL }

func readServerURL() (ret *url.URL, err error) {
	u := sdkclient.DefaultLocalURL
	if _, err = cfg.Get("http.service_url", &u); err != nil {
		return
	} // if not overriden by config, then url will remain default

	if ret, err = url.Parse(u); err != nil {
		return
	}

	if ret.Scheme == "" {
		ret.Scheme = "http"
	}

	return
}
