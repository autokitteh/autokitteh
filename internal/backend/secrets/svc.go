package secrets

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type Secrets interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	Delete(ctx context.Context, key string) error
}

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
	if cfg.Provider == "" {
		return nil, sdkerrors.NewInvalidArgumentError("secret provider cannot be empty")
	}

	var (
		provider Secrets
		err      error
	)

	switch cfg.Provider {
	case SecretProviderAWSSecretManager:
		provider, err = newAWSSecrets(z, cfg.AWSSecretManager)
	case SecretProviderDatabase:
		provider, err = newDatabaseSecrets(z, db)
	case SecretProviderVault:
		provider, err = newVaultSecrets(z, cfg.Vault)
	default:
		return nil, sdkerrors.NewInvalidArgumentError("invalid secret provider: %s", cfg.Provider)
	}

	if err != nil {
		return nil, sdkerrors.NewInvalidArgumentError("invalid configuration for %s secret provider", cfg.Provider)
	}

	return &secrets{
		cfg:      cfg,
		z:        z,
		provider: provider,
	}, nil

}

func (s *secrets) Get(ctx context.Context, key string) (string, error) {
	return s.provider.Get(ctx, wrapKey(s.cfg.GlobalScope, key))
}

func (s *secrets) Set(ctx context.Context, key string, value string) error {
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
