//go:build !enterprise
// +build !enterprise

package db

import "context"

type DB interface {
	Shared
}

type TX interface {
	DB

	Lock(ctx context.Context, id string) error
}
