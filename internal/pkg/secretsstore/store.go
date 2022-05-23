package secretsstore

import (
	"context"

	"go.autokitteh.dev/sdk/api/apiproject"
	"github.com/autokitteh/stores/pkvstore"
)

var ErrNotFound = pkvstore.ErrNotFound

type Store struct{ pkvstore.Store }

func (s *Store) Set(ctx context.Context, pid apiproject.ProjectID, name, v string) error {
	if v == "" {
		return s.Store.Delete(ctx, pid.String(), name)
	}

	return s.Store.Put(ctx, pid.String(), name, []byte(v))
}

func (s *Store) Get(ctx context.Context, pid apiproject.ProjectID, name string) (string, error) {
	bs, err := s.Store.Get(ctx, pid.String(), name)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func (s *Store) List(ctx context.Context, pid apiproject.ProjectID) ([]string, error) {
	return s.Store.List(ctx, pid.String())
}

func (s *Store) Setup(ctx context.Context) error { return s.Store.Setup(ctx) }
