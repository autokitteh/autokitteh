//go:build !enterprise

package db

type DB interface {
	Shared
}
