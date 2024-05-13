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
func newDatabaseSecrets(l *zap.Logger, db db.DB) (*dbSecrets, error) {
	return &dbSecrets{logger: l, db: db}, nil
}

func (s *dbSecrets) Set(ctx context.Context, key string, data string) error {
	return s.db.SetSecret(ctx, key, data)
}

func (s *dbSecrets) Get(ctx context.Context, key string) (string, error) {
	return s.db.GetSecret(ctx, key)
}

func (s *dbSecrets) Delete(ctx context.Context, key string) error {
	return s.db.DeleteSecret(ctx, key)
}
