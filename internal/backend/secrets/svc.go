package secrets

import (
	"context"
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.uber.org/zap"
)

type Secrets interface {
	Get(ctx context.Context, key string) (map[string]string, error)
	Set(ctx context.Context, key string, value map[string]string) error
	Delete(ctx context.Context, key string) error
}

var (
	ErrSecretsInvalidProvider              = errors.New("invalid provider")
	ErrSecretsInvalidProviderConfiguration = errors.New("invalid provider configuration")
)

const (
	SecretProviderDatabase         = "db"
	SecretProviderAWSSecretManager = "awsSecretManager"
	SecretProviderVault            = "vault"
)

type secrets struct {
	cfg      *Config
	z        *zap.Logger
	provider Secrets
}

func New(cfg *Config, z *zap.Logger, db db.DB) (Secrets, error) {
	var (
		provider Secrets
		err      error
	)

	switch cfg.Provider {
	case SecretProviderAWSSecretManager:
		if provider, err = newAWSSecrets(z, cfg); err != nil {
			z.Panic("invalid provider configuration", zap.Error(err))
			return nil, ErrSecretsInvalidProviderConfiguration
		}
	case SecretProviderDatabase:
		if provider, err = newDatabaseSecrets(z, db); err != nil {
			z.Panic("invalid provider configuration", zap.Error(err))
			return nil, ErrSecretsInvalidProviderConfiguration
		}
	}

	if provider == nil {
		return nil, ErrSecretsInvalidProvider
	}

	return &secrets{
		cfg:      cfg,
		z:        z,
		provider: provider,
	}, nil

}

func (s *secrets) Get(ctx context.Context, key string) (map[string]string, error) {
	return s.provider.Get(ctx, wrapKey(s.cfg.GlobalScope, key))
}

func (s *secrets) Set(ctx context.Context, key string, value map[string]string) error {
	return s.provider.Set(ctx, wrapKey(s.cfg.GlobalScope, key), value)
}

func (s *secrets) Delete(ctx context.Context, key string) error {
	return s.provider.Delete(ctx, wrapKey(s.cfg.GlobalScope, key))
}

func wrapKey(scope string, key string) string {
	if scope != "" {
		return fmt.Sprintf("%s/%s", scope, key)
	}

	return key
}
