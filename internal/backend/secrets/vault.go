package secrets

import (
	"context"
	"fmt"
	"time"

	vault "github.com/hashicorp/vault/api"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type vaultSecrets struct {
	client *vault.KVv2
	logger *zap.Logger
}

// NewVaultSecrets initializes a client connection to HashiCorp Vault
// (https://www.vaultproject.io/).
func NewVaultSecrets(l *zap.Logger) (Secrets, error) {
	config := vault.DefaultConfig()
	client, err := vault.NewClient(config)
	if err != nil {
		l.Error("Vault client initialization error",
			zap.Error(err),
		)
		return nil, err
	}
	s := &vaultSecrets{
		client: client.KVv2("secret"),
		logger: l,
	}
	return s, nil
}

// The data size limit is 0.5 or 1 MiB, according to this link:
// https://developer.hashicorp.com/vault/docs/internals/limits
func (s *vaultSecrets) Set(scope, name string, data map[string]string) error {
	namePath := secretPath(scope, name)
	d := kittehs.TransformMapValues(data, func(s string) any { return s })
	if _, err := s.client.Put(context.Background(), namePath, d); err != nil {
		return err
	}
	return nil
}

func (s *vaultSecrets) Get(scope, name string) (map[string]string, error) {
	sec, err := s.client.Get(context.Background(), secretPath(scope, name))
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
	namePath := secretPath(scope, name)
	data := map[string]any{token: time.Now().UTC().Format(time.RFC3339)}
	if _, err := s.client.Patch(context.Background(), namePath, data); err != nil {
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

func (s *vaultSecrets) Delete(scope, name string) error {
	namePath := secretPath(scope, name)
	if err := s.client.DeleteMetadata(context.Background(), namePath); err != nil {
		return err
	}
	return nil
}
