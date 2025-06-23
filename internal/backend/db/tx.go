package db

import "context"

// TX is the interface for operations allows in a database transaction.
type TX interface {
	DB

	Lock(ctx context.Context, id string) error
}
