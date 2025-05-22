//go:build enterprise
// +build enterprise

package dbgorm

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"gorm.io/gorm"
)

func (gdb *gormdb) GetWorkerInfo(ctx context.Context, id string) (scheme.WorkerInfo, error) {
	workerInfo := scheme.WorkerInfo{
		WorkerID: id,
	}
	if err := gdb.reader.WithContext(ctx).Where("worker_id = ?", id).First(&workerInfo).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return scheme.WorkerInfo{}, err
		}
		if err := gdb.writer.WithContext(ctx).Create(&workerInfo).Error; err != nil {
			return scheme.WorkerInfo{}, err
		}
	}
	return workerInfo, nil
}

func (gdb *gormdb) IncActiveWorkflows(ctx context.Context, workerID string) (int, error) {
	var currentActiveWorkflows int
	if err := gdb.writer.WithContext(ctx).Raw("UPDATE worker_infos SET active_workflows = active_workflows + 1, updated_at = NOW() WHERE worker_id = $1 RETURNING active_workflows", workerID).Scan(&currentActiveWorkflows).Error; err != nil {
		return 0, err
	}

	return currentActiveWorkflows, nil
}
func (gdb *gormdb) DecActiveWorkflows(ctx context.Context, workerID string) (int, error) {
	var currentActiveWorkflows int
	if err := gdb.writer.WithContext(ctx).Raw("UPDATE worker_infos SET active_workflows = GREATEST(active_workflows - 1, 0), updated_at = NOW() WHERE worker_id = $1 RETURNING active_workflows", workerID).Scan(&currentActiveWorkflows).Error; err != nil {
		return 0, err
	}
	return currentActiveWorkflows, nil
}
