package dbgorm

import (
	"context"
	"time"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) createDeployment(ctx context.Context, deployment *scheme.Deployment) error {
	return db.db.WithContext(ctx).Create(deployment).Error
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

	return db.locked(func(db *gormdb) error {
		return translateError(db.createDeployment(ctx, &d))
	})
}

func (db *gormdb) getDeployment(ctx context.Context, deploymentID string) (*scheme.Deployment, error) {
	return getOne(db.db, ctx, scheme.Deployment{}, "deployment_id = ?", deploymentID)
}

func (db *gormdb) GetDeployment(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	d, err := db.getDeployment(ctx, id.String())
	if d == nil || err != nil {
		return sdktypes.InvalidDeployment, translateError(err)
	}
	return scheme.ParseDeployment(*d)
}

func (db *gormdb) deleteDeployment(ctx context.Context, deploymentID string) error {
	return db.deleteDeploymentsAndDependents(ctx, []string{deploymentID})
}

// delete deployments and relevant sessions
func (db *gormdb) deleteDeploymentsAndDependents(ctx context.Context, depIDs []string) error {
	// NOTE: should be transactional

	if len(depIDs) == 0 {
		return nil
	}
	gormDB := db.db.WithContext(ctx)

	if err := gormDB.Delete(&scheme.Session{}, "deployment_id IN ?", depIDs).Error; err != nil {
		return err
	}

	// FIXME: do not delete deployment
	// buildIDs := gormDB.Model(&scheme.Deployment{}).Select("build_id").Where("deployment_id IN (?)", depIDs)
	// if err := gormDB.Delete(&scheme.Build{}, "build_id IN (?)", buildIDs).Error; err != nil {
	// 	return err
	// }

	return gormDB.Delete(&scheme.Deployment{}, "deployment_id IN ?", depIDs).Error
}

func (db *gormdb) DeleteDeployment(ctx context.Context, deploymentID sdktypes.DeploymentID) error {
	return db.transaction(ctx, func(tx *tx) error {
		return translateError(tx.deleteDeployment(ctx, deploymentID.String()))
	})
}

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
	COUNT(case when sessions.current_state_type = ? then 1 end) AS completed,
	COUNT(case when sessions.current_state_type = ? then 1 end) AS stopped
	`, sdktypes.SessionStateTypeCreated.ToProto(),
		sdktypes.SessionStateTypeRunning.ToProto(),
		sdktypes.SessionStateTypeError.ToProto(),
		sdktypes.SessionStateTypeCompleted.ToProto(),
		sdktypes.SessionStateTypeStopped.ToProto()).
		Joins(`LEFT JOIN sessions on deployments.deployment_id = sessions.deployment_id
	AND sessions.deleted_at IS NULL`).
		Group("deployments.deployment_id")

	var ds []scheme.DeploymentWithStats
	if err := q.Find(&ds).Error; err != nil {
		return nil, err
	}
	return ds, nil
}

func (db *gormdb) listDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]scheme.Deployment, error) {
	q := db.listDeploymentsCommonQuery(ctx, filter)
	var ds []scheme.Deployment
	if err := q.Find(&ds).Error; err != nil {
		return nil, err
	}
	return ds, nil
}

func (db *gormdb) ListDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
	if filter.IncludeSessionStats {
		ds, err := db.listDeploymentsWithStats(ctx, filter)
		if ds == nil {
			return nil, translateError(err)
		}
		return kittehs.TransformError(ds, scheme.ParseDeploymentWithSessionStats)
	} else {
		ds, err := db.listDeployments(ctx, filter)
		if ds == nil {
			return nil, translateError(err)
		}
		return kittehs.TransformError(ds, scheme.ParseDeployment)
	}
}

func (db *gormdb) updateDeploymentState(ctx context.Context, id string, state sdktypes.DeploymentState) error {
	d := &scheme.Deployment{DeploymentID: id}

	return db.locked(func(db *gormdb) error {
		result := db.db.WithContext(ctx).Model(d).Updates(
			map[string]any{"state": state.ToProto(), "updated_at": time.Now()})
		if result.Error != nil {
			return translateError(result.Error)
		}
		if result.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return nil
	})
}

func (db *gormdb) UpdateDeploymentState(ctx context.Context, id sdktypes.DeploymentID, state sdktypes.DeploymentState) error {
	return db.updateDeploymentState(ctx, id.String(), state)
}
