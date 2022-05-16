package kvstore

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Put(context.Context, string, []byte) error
	Delete(context.Context, string) error
	Get(context.Context, string) ([]byte, error)

	Setup(context.Context) error
	Teardown(context.Context) error
}
