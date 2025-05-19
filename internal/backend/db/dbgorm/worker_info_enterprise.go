//go:build enterprise
// +build enterprise

package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (gdb *gormdb) GetWorkerInfo(ctx context.Context, id string) (scheme.WorkerInfo, error) {
	return scheme.WorkerInfo{}, nil
}
