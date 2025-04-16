package dbgorm

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"gorm.io/gorm"
)

func (gdb *gormdb) AddJob(ctx context.Context, jobType scheme.JobType, data map[string]any) (uuid.UUID, error) {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return uuid.Nil, err
	}

	job := scheme.Job{
		JobID:  uuid.New(),
		Type:   jobType,
		Status: scheme.JobStatusPending,
		Data:   dataJson,
	}

	if err := gdb.writer.WithContext(ctx).Create(&job).Error; err != nil {
		return uuid.Nil, err
	}

	return job.JobID, nil
}

func (gdb *gormdb) GetPendingJobs(ctx context.Context, count int) ([]scheme.Job, error) {
	jobs := make([]scheme.Job, count)

	// Need to think if here is the right place to set the default value
	// for count. It should be set in the caller, but we need to make sure values
	// are valid. maybe we should error on invalid count value.
	if count > 10 {
		count = 10
	}
	if count <= 0 {
		count = 1
	}

	//TODO: Need to fetch pending jobs or jobs that are in status acquired for too long -
	// Acquired means the worker poll the job, and still havent processed it yet
	// In temporal sense, it means it didn't call executeWorkflow on the message yet
	// For future non temporal workflow, it means it didn't finish executing the workflow yet
	// In this cases, we need to retry the job on another worker after some grace period
	// This is to prevent the job from being stuck in acquired state forever
	if err := gdb.writer.WithContext(ctx).Raw(`
	UPDATE jobs
	SET status = '$1', retry_count = retry_count + 1, updated_at = NOW(), start_processing_time = NOW()
	WHERE id IN (SELECT id FROM jobs WHERE status = 'pending' LIMIT $2 FOR UPDATE SKIP LOCKED)
	RETURNING *;
	`, scheme.JobStatusAcquired, count).Scan(&jobs).Error; err != nil {
		return nil, err
	}

	return jobs, nil
}

func (gdb *gormdb) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status scheme.JobStatus) error {
	q := gdb.writer.WithContext(ctx).Model(&scheme.Job{}).Where("id = ?", jobID).UpdateColumn("status", status).UpdateColumn("updated_at", time.Now())

	if status == scheme.JobStatusDone || status == scheme.JobStatusFailed {
		q = q.UpdateColumn("end_processing_time", gorm.Expr("NOW()"))
	}

	if err := q.Error; err != nil {
		return err
	}

	return nil
}
