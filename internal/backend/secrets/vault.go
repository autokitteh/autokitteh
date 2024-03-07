package secrets

import (
	"fmt"
	"time"

	vault "github.com/hashicorp/vault/api"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type vaultSecrets struct {
	client  *vault.KVv2
	logger  *zap.Logger
	timeout time.Duration
}

// NewVaultSecrets initializes a client connection to HashiCorp Vault.
func NewVaultSecrets(l *zap.Logger, c *Config) (Secrets, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = c.VaultURL

	client, err := vault.NewClient(cfg)
	if err != nil {
		l.Error("Vault client initialization", zap.Error(err))
		return nil, err
	}

	return &vaultSecrets{client: client.KVv2("secret"), logger: l, timeout: c.Timeout}, nil
}

// The data size limit is 0.5 or 1 MiB, according to this link:
// https://developer.hashicorp.com/vault/docs/internals/limits
func (s *vaultSecrets) Set(scope, name string, data map[string]string) error {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	d := kittehs.TransformMapValues(data, func(s string) any { return s })
	if _, err := s.client.Put(ctx, secretPath(scope, name), d); err != nil {
		return err
	}
	return nil
}

func (s *vaultSecrets) Get(scope, name string) (map[string]string, error) {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	sec, err := s.client.Get(ctx, secretPath(scope, name))
	if err != nil {
		return nil, err
	}

	data, err := kittehs.TransformMapValuesError(sec.Data, any2str)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *vaultSecrets) Append(scope, name, token string) error {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	data := map[string]any{token: time.Now().UTC().Format(time.RFC3339)}
	if _, err := s.client.Patch(ctx, secretPath(scope, name), data); err != nil {
		return err
	}
	return nil
}

func (s *vaultSecrets) Delete(scope, name string) error {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	if err := s.client.DeleteMetadata(ctx, secretPath(scope, name)); err != nil {
		return err
	}
	return nil
}

func any2str(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string: %v", v)
	}
	return s, nil
}
