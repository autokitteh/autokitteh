//go:build !enterprise
// +build !enterprise

package db

type DB interface {
	Shared
}
