package secrets

import (
	"go.uber.org/zap"
)

// NewFakeSecrets initializes a fake secrets manager for unit-testing.
// It's similar to [NewFileSecrets], but entirely in-memory.
func NewFakeSecrets(l *zap.Logger) (Secrets, error) {
	s := &fileSecrets{
		secrets: map[string]map[string]string{},
		logger:  l,
	}
	l.Warn("Using an ephemeral in-memory secrets manager")
	return s, nil
}
