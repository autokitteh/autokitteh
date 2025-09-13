package db

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TX is the interface for operations allows in a database transaction.
type TX interface {
	DB

	LockProject(ctx context.Context, pid sdktypes.ProjectID) error
}
