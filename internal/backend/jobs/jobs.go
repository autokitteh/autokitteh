package jobs

import (
	"context"

	"github.com/google/uuid"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

type JobManager struct {
	z  *zap.Logger
	db db.DB
}

func New(l *zap.Logger, db db.DB) *JobManager {
	return &JobManager{
		z:  l.Named("jobs"),
		db: db,
	}
}

func (jm *JobManager) StartSession(ctx context.Context, sessoin sdktypes.Session) error {
	data := map[string]any{
		"session_id": sessoin.ID(),
	}
	_, err := jm.db.AddJob(ctx, scheme.JobTypeRun, data)
	return err
}

func (jm *JobManager) GetPendingSessionJob(ctx context.Context) (*scheme.Job, error) {
	jobs, err := jm.db.GetPendingJobs(ctx, 1)
	if err != nil {
		return nil, err
	}

	if len(jobs) == 0 {
		return nil, nil

	}

	return &jobs[0], nil
}

func (jm *JobManager) MarkJobDone(ctx context.Context, jobID uuid.UUID) error {
	return jm.db.UpdateJobStatus(ctx, jobID, scheme.JobStatusDone)
}

func (jm *JobManager) MarkJobFailed(ctx context.Context, jobID uuid.UUID) error {
	return jm.db.UpdateJobStatus(ctx, jobID, scheme.JobStatusFailed)
}
