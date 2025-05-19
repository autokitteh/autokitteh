//go:build enterprise
// +build enterprise

package db

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

type DB interface {
	Shared
	GetWorkerInfo(ctx context.Context, id string) (scheme.WorkerInfo, error)
	UpdateWorkerInfo(ctx context.Context, id string, info scheme.WorkerInfo) error
}
