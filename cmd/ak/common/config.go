package common

import (
	"path/filepath"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

const (
	EnvVarPrefix       = "AK_"
	ConfigYAMLFileName = "config.yaml"
)

var cfg *svc.Config

func ConfigYAMLFilePath() string {
	return filepath.Join(xdg.ConfigHomeDir(), ConfigYAMLFileName)
}

func InitConfig(confmap map[string]any) (err error) {
	// Reminder: cfg is a package-scoped variable, not function-scoped.
	cfg, err = svc.LoadConfig(EnvVarPrefix, confmap, ConfigYAMLFilePath())
	return
}

func Config() *svc.Config { return cfg }
