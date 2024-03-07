package secrets

import (
	"context"
	"fmt"
	"time"

	"github.com/lithammer/shortuuid"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Config struct {
	Type     string        `koanf:"type"`
	VaultURL string        `koanf:"vault_url"`
	Timeout  time.Duration `koanf:"timeout_duration"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}

var defaultTimeout = 1 * time.Minute

func New(l *zap.Logger, cfg *Config) (sdkservices.Secrets, error) {
	var impl Secrets
	var err error

	switch cfg.Type {
	case "aws":
		impl, err = NewAWSSecrets(l, cfg)
	// TODO: case "db"
	case "vault":
		impl, err = NewVaultSecrets(l, cfg)
	default:
		impl, err = NewFileSecrets(l, xdg.DataHomeDir())
	}

	if err != nil {
		return nil, err
	}

	return &secrets{impl: impl, logger: l}, nil
}

type secrets struct {
	impl   Secrets
	logger *zap.Logger
}

func (s secrets) Create(ctx context.Context, scope string, data map[string]string, key string) (string, error) {
	l := s.logger.With(zap.String("scope", scope))
	token := newConnectionToken()

	// Connection token --> OAUth token, etc. (to call API methods).
	name := connectionSecretName(token)
	if err := s.impl.Set(scope, name, data); err != nil {
		l.Error("Failed to save connection",
			zap.String("secretName", name),
			zap.Error(err),
		)
		return "", err
	}

	// Integration-specific key --> connection token(s) (to dispatch API events).
	if err := s.impl.Append(scope, key, token); err != nil {
		l.Error("Failed to save reverse connection mapping",
			zap.String("secretName", key),
			zap.Error(err),
		)
		name = connectionSecretName(token)
		if err := s.impl.Delete(scope, name); err != nil {
			l.Error("Dangling connection mapping",
				zap.String("secretName", name),
				zap.Error(err),
			)
		}
		return "", err
	}

	// Success.
	return token, nil
}

func (s secrets) Get(ctx context.Context, scope, token string) (map[string]string, error) {
	l := s.logger.With(zap.String("integration", scope))

	name := connectionSecretName(token)
	data, err := s.impl.Get(scope, name)
	if err != nil {
		l.Error("Failed to load connection",
			zap.String("secretName", name),
			zap.Error(err),
		)
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("secret not found: %s", name)
	}

	// Success.
	return data, nil
}

func (s secrets) List(ctx context.Context, scope, key string) ([]string, error) {
	l := s.logger.With(zap.String("scope", scope))

	data, err := s.impl.Get(scope, key)
	if err != nil {
		l.Error("Failed to list connections",
			zap.String("secretName", key),
			zap.Error(err),
		)
		return nil, err
	}
	if data == nil {
		data = map[string]string{}
	}

	// Success.
	return maps.Keys(data), nil
}

func limitedContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func newConnectionToken() string {
	s := shortuuid.New()
	// Separators help users check and compare strings visually.
	// "_" doesn't interfere with double-click selection like "-".
	return fmt.Sprintf("%s_%s_%s", s[0:7], s[7:15], s[15:])
}

func connectionSecretName(token string) string {
	return "connections/" + token
}
