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
	IncActiveWorkflows(ctx context.Context, workerID string) (int, error)
	DecActiveWorkflows(ctx context.Context, workerID string) (int, error)
}
