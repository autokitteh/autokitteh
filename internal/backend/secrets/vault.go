package secrets

import (
	"context"
	"errors"

	vault "github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

type vaultSecrets struct {
	client *vault.KVv2
	logger *zap.Logger
}

// NewVaultSecrets initializes a client connection to Vault.
func newVaultSecrets(l *zap.Logger, c *vaultConfig) (*vaultSecrets, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = c.URL

	client, err := vault.NewClient(cfg)
	if err != nil {
		l.Error("Vault client initialization", zap.Error(err))
		return nil, err
	}

	return &vaultSecrets{client: client.KVv2("secret"), logger: l}, nil
}

// The data size limit is 0.5 or 1 MiB, according to this link:
// https://developer.hashicorp.com/vault/docs/internals/limits
func (s *vaultSecrets) Set(ctx context.Context, key string, value string) error {
	_, err := s.client.Put(ctx, key, map[string]any{"value": value})
	return err
}

func (s *vaultSecrets) Get(ctx context.Context, key string) (string, error) {
	sec, err := s.client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	data, ok := sec.Data["value"].(string)
	if !ok {
		return "", errors.New("invalid data")
	}
	return data, nil
}

func (s *vaultSecrets) Delete(ctx context.Context, key string) error {
	return s.client.DeleteMetadata(ctx, key)
}
