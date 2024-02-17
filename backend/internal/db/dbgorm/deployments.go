package dbgorm

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) CreateDeployment(ctx context.Context, deployment sdktypes.Deployment) error {
	now := time.Now()

	e := scheme.Deployment{
		DeploymentID: sdktypes.GetDeploymentID(deployment).String(),
		BuildID:      sdktypes.GetDeploymentBuildID(deployment).String(),
		EnvID:        sdktypes.GetDeploymentEnvID(deployment).String(),
		State:        int32(sdktypes.GetDeploymentState(deployment)),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Assuming if buildID already exists, nothing will happen and no error
	if err := db.locked(func(db *gormdb) error {
		return db.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&e).Error
	}); err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) GetDeployment(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	return get(db.db, ctx, scheme.ParseDeployment, "deployment_id = ?", id.String())
}

func (db *gormdb) DeleteDeployment(ctx context.Context, id sdktypes.DeploymentID) error {
	var b scheme.Deployment
	if err := db.db.WithContext(ctx).Where("deployment_id = ?", id.String()).Delete(&b).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) ListDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
	q := db.db.WithContext(ctx).Model(&scheme.Deployment{})
	if filter.BuildID != nil {
		q = q.Where("deployments.build_id = ?", filter.BuildID.String())
	}
	if filter.EnvID != nil {
		q = q.Where("deployments.env_id = ?", filter.EnvID.String())
	}
	if filter.State != sdktypes.DeploymentStateUnspecified {
		q = q.Where("deployments.state = ?", int32(filter.State))
	}

	if filter.Limit > 0 {
		q = q.Limit(int(filter.Limit))
	}

	q = q.Order("created_at desc")

	if filter.IncludeSessionStats {
		q = q.Select(`
			deployments.*, 
			count(case when sessions.current_state_type = ? then 1 end) as created,
			count(case when sessions.current_state_type = ? then 1 end) as running,
			count(case when sessions.current_state_type = ? then 1 end) as error,
			count(case when sessions.current_state_type = ? then 1 end) as completed
		`, sdktypes.CreatedSessionStateType,
			sdktypes.RunningSessionStateType,
			sdktypes.ErrorSessionStateType,
			sdktypes.CompletedSessionStateType).
			Joins("left join sessions on deployments.deployment_id = sessions.deployment_id").
			Group("deployments.deployment_id")
		var ds []scheme.DeploymentWithStats
		if err := q.Find(&ds).Error; err != nil {
			return nil, translateError(err)
		}

		return kittehs.TransformError(ds, scheme.ParseDeploymentWithSessionStats)
	}

	var ds []scheme.Deployment
	if err := q.Find(&ds).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(ds, scheme.ParseDeployment)
}

func (db *gormdb) UpdateDeploymentState(ctx context.Context, id sdktypes.DeploymentID, state sdktypes.DeploymentState) error {
	d := &scheme.Deployment{DeploymentID: id.String()}

	return db.locked(func(db *gormdb) error {
		result := db.db.WithContext(ctx).Model(d).Updates(map[string]any{"state": int32(state), "updated_at": time.Now()})
		if result.Error != nil {
			return translateError(result.Error)
		}
		if result.RowsAffected == 0 {
			return sdkerrors.ErrNotFound
		}

		return nil
	})
}
