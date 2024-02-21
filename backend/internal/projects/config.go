package projects

import (
	"go.autokitteh.dev/autokitteh/backend/configset"
)

type ResourcesConfig struct {
	AllowLocalFS bool `koanf:"allow_local_fs"`
}

type Config struct {
	Resources ResourcesConfig `koanf:"resources"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		Resources: ResourcesConfig{
			AllowLocalFS: true,
		},
	},
}
