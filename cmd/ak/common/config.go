package common

import (
	"path/filepath"

	"go.autokitteh.dev/autokitteh/backend/basesvc"
	"go.autokitteh.dev/autokitteh/backend/config"
)

const (
	EnvVarPrefix       = "AK_"
	ConfigYAMLFileName = "config.yaml"
)

var cfg *basesvc.Config

func ConfigYAMLFilePath() string {
	return filepath.Join(config.ConfigHomeDir(), ConfigYAMLFileName)
}

func InitConfig(confmap map[string]any) (err error) {
	// Reminder: cfg is a package-scoped variable, not function-scoped.
	cfg, err = basesvc.LoadConfig(EnvVarPrefix, confmap, ConfigYAMLFilePath())
	return
}

func Config() *basesvc.Config { return cfg }
