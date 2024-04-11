package secrets

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
)

type dbSecrets struct {
	logger *zap.Logger
	db     db.DB
}

// NewDatabaseSecrets initializes a (simple and persistent, yet insecure)
// secrets manager for local non-production usage, in AK's relational database.
// DO NOT STORE REAL SECRETS IN THIS WAY FOR LONG PERIODS OF TIME!
func NewDatabaseSecrets(l *zap.Logger, db db.DB) (Secrets, error) {
	return &dbSecrets{logger: l, db: db}, nil
}

func (s *dbSecrets) Set(ctx context.Context, scope, name string, data map[string]string) error {
	return s.db.SetSecret(ctx, secretPath(scope, name), data)
}

func (s *dbSecrets) Get(ctx context.Context, scope, name string) (map[string]string, error) {
	return s.db.GetSecret(ctx, secretPath(scope, name))
}

func (s *dbSecrets) Append(ctx context.Context, scope, name, token string) error {
	return s.db.AppendSecret(ctx, secretPath(scope, name), token)
}

func (s *dbSecrets) Delete(ctx context.Context, scope, name string) error {
	return s.db.DeleteSecret(ctx, secretPath(scope, name))
}
