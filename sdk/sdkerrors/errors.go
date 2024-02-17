package sdkerrors

import (
	"errors"
)

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrRPC             = errors.New("rpc")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrConflict        = errors.New("conflict")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrUnknown         = errors.New("unknown")
)

func IgnoreNotFoundErr[T any](t *T, err error) (*T, error) {
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	if t == nil {
		return nil, nil
	}

	return t, nil
}
