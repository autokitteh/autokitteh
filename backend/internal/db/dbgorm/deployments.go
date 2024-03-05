package dbgorm

import (
	"context"
	"time"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) createDeployment(ctx context.Context, deployment scheme.Deployment) error {
	return db.db.WithContext(ctx).Create(&deployment).Error
}

func (db *gormdb) CreateDeployment(ctx context.Context, deployment sdktypes.Deployment) error {
	now := time.Now()

	d := scheme.Deployment{
		DeploymentID: deployment.ID().String(),
		BuildID:      deployment.BuildID().String(),
		EnvID:        deployment.EnvID().String(),
		State:        int32(deployment.State().ToProto()),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := db.locked(func(db *gormdb) error {
		return db.createDeployment(ctx, d)
	}); err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) getDeployment(ctx context.Context, deploymentID string) (*scheme.Deployment, error) {
	return getOne(db.db, ctx, scheme.Deployment{}, "deployment_id = ?", deploymentID)
}

func (db *gormdb) GetDeployment(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	d, err := db.getDeployment(ctx, id.String())
	if d == nil || err != nil {
		return sdktypes.InvalidDeployment, err
	}
	return scheme.ParseDeployment(*d)
}

func (db *gormdb) DeleteDeployment(ctx context.Context, id sdktypes.DeploymentID) error {
	var b scheme.Deployment
	if err := db.db.WithContext(ctx).Where("deployment_id = ?", id.String()).Delete(&b).Error; err != nil {
		return translateError(err)
	}
	// FIXME: delete deployment sessions as well?

	return nil
}

// FIXME: fix generic in order to avoid all this
func (db *gormdb) listDeploymentsCommonQuery(ctx context.Context, filter sdkservices.ListDeploymentsFilter) *gorm.DB {
	q := db.db.WithContext(ctx).Model(&scheme.Deployment{})
	if filter.BuildID.IsValid() {
		q = q.Where("deployments.build_id = ?", filter.BuildID.String())
	}
	if filter.EnvID.IsValid() {
		q = q.Where("deployments.env_id = ?", filter.EnvID.String())
	}
	if filter.State != sdktypes.DeploymentStateUnspecified {
		q = q.Where("deployments.state = ?", filter.State.ToProto())
	}

	if filter.Limit > 0 {
		q = q.Limit(int(filter.Limit))
	}

	q = q.Order("created_at desc")
	return q
}

func (db *gormdb) listDeploymentsWithStats(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]scheme.DeploymentWithStats, error) {
	q := db.listDeploymentsCommonQuery(ctx, filter)

	q = q.Select(`
	deployments.*, 
	COUNT(case when sessions.current_state_type = ? then 1 end) AS created,
	COUNT(case when sessions.current_state_type = ? then 1 end) AS running,
	COUNT(case when sessions.current_state_type = ? then 1 end) AS error,
	COUNT(case when sessions.current_state_type = ? then 1 end) AS completed
	`, sdktypes.SessionStateTypeCreated.ToProto(),
		sdktypes.SessionStateTypeRunning.ToProto(),
		sdktypes.SessionStateTypeError.ToProto(),
		sdktypes.SessionStateTypeCompleted.ToProto()).
		Joins(`LEFT JOIN sessions on deployments.deployment_id = sessions.deployment_id
	AND sessions.deleted_at IS NULL`).
		Group("deployments.deployment_id")

	var ds []scheme.DeploymentWithStats
	if err := q.Find(&ds).Error; err != nil {
		return nil, translateError(err)
	}
	return ds, nil
}

func (db *gormdb) listDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]scheme.Deployment, error) {
	q := db.listDeploymentsCommonQuery(ctx, filter)
	var ds []scheme.Deployment
	if err := q.Find(&ds).Error; err != nil {
		return nil, translateError(err)
	}
	return ds, nil
}

func (db *gormdb) ListDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
	if filter.IncludeSessionStats {
		ds, err := db.listDeploymentsWithStats(ctx, filter)
		if ds == nil {
			return nil, err
		}
		return kittehs.TransformError(ds, scheme.ParseDeploymentWithSessionStats)
	} else {
		ds, err := db.listDeployments(ctx, filter)
		if ds == nil {
			return nil, err
		}
		return kittehs.TransformError(ds, scheme.ParseDeployment)
	}
}

func (db *gormdb) UpdateDeploymentState(ctx context.Context, id sdktypes.DeploymentID, state sdktypes.DeploymentState) error {
	d := &scheme.Deployment{DeploymentID: id.String()}

	return db.locked(func(db *gormdb) error {
		result := db.db.WithContext(ctx).Model(d).Updates(map[string]any{"state": state.ToProto(), "updated_at": time.Now()})
		if result.Error != nil {
			return translateError(result.Error)
		}
		if result.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return nil
	})
}
