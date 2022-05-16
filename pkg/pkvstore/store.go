package pkvstore

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	Put(context.Context, string, string, []byte) error
	Delete(context.Context, string, string) error
	Get(context.Context, string, string) ([]byte, error)
	List(context.Context, string) ([]string, error)

	Setup(context.Context) error
	Teardown(context.Context) error
}
