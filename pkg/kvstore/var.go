package kvstore

import (
	"context"
)

type VarStore interface {
	Set(context.Context, []byte) error
	Get(context.Context) ([]byte, error)
}

type varStore struct {
	s Store
	n string
}

func NewVarStore(s Store, n string) VarStore { return &varStore{s: s, n: n} }

func (s *varStore) Set(ctx context.Context, v []byte) error { return s.s.Put(ctx, s.n, v) }
func (s *varStore) Get(ctx context.Context) ([]byte, error) { return s.s.Get(ctx, s.n) }
