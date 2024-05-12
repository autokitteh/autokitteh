package secrets

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type awsSecretManagerConfig struct{}

type Config struct {
	Provider         string                  `koanf:"provider"`
	GlobalScope      string                  `koanf:"global_scope"`
	AWSSecretManager *awsSecretManagerConfig `koanf:"aws"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{},
		Dev: &Config{
			Provider: SecretProviderDatabase,
		},
	}
)
