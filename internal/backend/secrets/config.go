package secrets

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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

func (c Config) Validate() error {
	if c.Provider != "" && !kittehs.ContainedIn(providers...)(c.Provider) {
		return fmt.Errorf("invalid secret provider: %s", c.Provider)
	}

	return nil
}

var Configs = config.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		Provider: SecretProviderDatabase,
	},
}
