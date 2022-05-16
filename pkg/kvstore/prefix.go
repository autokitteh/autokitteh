package kvstore

import (
	"context"
	"errors"
)

type StoreWithKeyPrefix struct {
	Store  Store
	Prefix string
}

var _ Store = &StoreWithKeyPrefix{}

func (s *StoreWithKeyPrefix) Put(ctx context.Context, k string, v []byte) error {
	return s.Store.Put(ctx, s.Prefix+k, v)
}

func (s *StoreWithKeyPrefix) Delete(ctx context.Context, k string) error {
	return s.Store.Delete(ctx, s.Prefix+k)
}

func (s *StoreWithKeyPrefix) Get(ctx context.Context, k string) ([]byte, error) {
	return s.Store.Get(ctx, s.Prefix+k)
}

func (s *StoreWithKeyPrefix) Setup(ctx context.Context) error {
	return errors.New("not implenented")
}

func (s *StoreWithKeyPrefix) Teardown(ctx context.Context) error {
	return errors.New("not implemnented")
}
