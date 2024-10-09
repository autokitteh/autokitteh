package secrets

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type awsSecretManagerConfig struct{}

type vaultConfig struct {
	URL string `koanf:"url"`
}

type Config struct {
	Provider         string                  `koanf:"provider"`
	GlobalScope      string                  `koanf:"global_scope"`
	AWSSecretManager *awsSecretManagerConfig `koanf:"aws"`
	Vault            *vaultConfig            `koanf:"vault"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{},
		Dev: &Config{
			Provider: SecretProviderDatabase,
		},
	}
)
