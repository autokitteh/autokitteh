package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) createDeployment(ctx context.Context, d *scheme.Deployment) error {
	return gormErrNotFoundToForeignKey(gdb.writer.WithContext(ctx).Create(d).Error)
}

func (gdb *gormdb) deleteDeployment(ctx context.Context, deploymentID uuid.UUID) error {
	return gdb.writeTransaction(ctx, func(tx *gormdb) error {
		return tx.deleteDeploymentsAndDependents(ctx, []uuid.UUID{deploymentID})
	})
}

// delete deployments and relevant sessions
func (gdb *gormdb) deleteDeploymentsAndDependents(ctx context.Context, depIDs []uuid.UUID) error {
	// NOTE: should be transactional

	if len(depIDs) == 0 {
		return nil
	}
	db := gdb.writer.WithContext(ctx)

	if err := db.Delete(&scheme.Session{}, "deployment_id IN ?", depIDs).Error; err != nil {
		return err
	}

	// FIXME: do not delete builds?
	// buildIDs := gormDB.Model(&scheme.Deployment{}).Select("build_id").Where("deployment_id IN (?)", depIDs)
	// if err := gormDB.Delete(&scheme.Build{}, "build_id IN (?)", buildIDs).Error; err != nil {
	// 	return err
	// }

	return db.Delete(&scheme.Deployment{}, "deployment_id IN ?", depIDs).Error
}

func (gdb *gormdb) updateDeploymentState(
	ctx context.Context,
	deploymentID uuid.UUID,
	state sdktypes.DeploymentState,
) (sdktypes.DeploymentState, error) {
	// NOTES:
	// - we want to return the old state of the deployment before it was updated.
	// - using RETURNING clause in UPDATE will return the new value, not the old one
	// - in postgres it's possible to use `UPDATE tbl t1 ... FROM tbl t2.. RETURNING t2.column`,
	//   e.g. effectively returning the old value before being updated). Unfortunately it won't work
	//   in SQLite, since RETURNING could use only canonical table name, and not the alias.
	// - that's why we are using two queries: one to get the old value, and the other to update it.

	d := scheme.Deployment{DeploymentID: deploymentID}
	var oldState int32 = 0

	if err := gdb.writeTransaction(ctx, func(tx *gormdb) error {
		if err := tx.writer.Model(&d).Select("state").First(&oldState).Error; err != nil {
			return err
		}

		data := updatedBaseColumns(ctx)
		data["state"] = int32(state.ToProto())

		if err := tx.writer.Model(&d).UpdateColumns(data).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return sdktypes.DeploymentStateUnspecified, err
	}
	state = kittehs.Must1(sdktypes.DeploymentStateFromProto(deploymentsv1.DeploymentState(oldState)))
	return state, nil
}

func (gdb *gormdb) getDeployment(ctx context.Context, deploymentID uuid.UUID) (*scheme.Deployment, error) {
	return getOne[scheme.Deployment](gdb.reader.WithContext(ctx), "deployment_id = ?", deploymentID)
}

func (gdb *gormdb) listDeploymentsCommonQuery(ctx context.Context, filter sdkservices.ListDeploymentsFilter) *gorm.DB {
	q := gdb.reader.WithContext(ctx)

	q = withProjectID(q, "deployments", filter.ProjectID)

	q = withProjectOrgID(q, filter.OrgID, "deployments")

	if filter.BuildID.IsValid() {
		q = q.Where("deployments.build_id = ?", filter.BuildID.UUIDValue())
	}
	if filter.ProjectID.IsValid() {
		q = q.Where("deployments.project_id = ?", filter.ProjectID.UUIDValue())
	}
	if filter.State != sdktypes.DeploymentStateUnspecified {
		q = q.Where("deployments.state = ?", filter.State.ToProto())
	}

	if filter.Limit > 0 {
		q = q.Limit(int(filter.Limit))
	}

	return q.Order("deployments.deployment_id desc")
}

func (db *gormdb) listDeploymentsWithStats(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]scheme.DeploymentWithStats, error) {
	q := db.listDeploymentsCommonQuery(ctx, filter)

	q = q.Model(scheme.Deployment{}).Select(`
        deployments.*,
        session_stats.created,
        session_stats.running,
        deployment_stats.completed,
        deployment_stats.error,
        deployment_stats.stopped
    `).Joins(`
        LEFT JOIN (
            SELECT 
                deployment_id,
                COUNT(CASE WHEN current_state_type = ? THEN 1 END) AS created,
                COUNT(CASE WHEN current_state_type = ? THEN 1 END) AS running
            FROM sessions
			WHERE current_state_type IN (?, ?)
            GROUP BY deployment_id
        ) AS session_stats ON session_stats.deployment_id = deployments.deployment_id
    `, int32(sdktypes.SessionStateTypeCreated.ToProto()),
		int32(sdktypes.SessionStateTypeRunning.ToProto()),
		int32(sdktypes.SessionStateTypeCreated.ToProto()),
		int32(sdktypes.SessionStateTypeRunning.ToProto())).
		Joins(`
        LEFT JOIN (
            SELECT 
                deployment_id,
                SUM(CASE WHEN session_state = ? THEN count ELSE 0 END) AS completed,
                SUM(CASE WHEN session_state = ? THEN count ELSE 0 END) AS error,
                SUM(CASE WHEN session_state = ? THEN count ELSE 0 END) AS stopped
            FROM deployment_session_stats
            GROUP BY deployment_id
        ) AS deployment_stats ON deployment_stats.deployment_id = deployments.deployment_id
    `, int32(sdktypes.SessionStateTypeCompleted.ToProto()),
			int32(sdktypes.SessionStateTypeError.ToProto()),
			int32(sdktypes.SessionStateTypeStopped.ToProto()))

	var ds []scheme.DeploymentWithStats
	if err := q.Find(&ds).Error; err != nil {
		return nil, err
	}
	return ds, nil
}

func (gdb *gormdb) incDeploymentStats(tx *gormdb, deploymentID uuid.UUID, stateType int) error {
	stats := scheme.DeploymentSessionStats{
		DeploymentID: deploymentID,
		SessionState: stateType,
		Count:        1,
	}

	return tx.writer.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "deployment_id"}, {Name: "session_state"}},
		DoUpdates: clause.Assignments(map[string]any{
			"count": gorm.Expr("deployment_session_stats.count + ?", 1),
		}),
	}).Create(&stats).Error
}

func (tx *gormdb) decDeploymentStats(ctx context.Context, session scheme.Session) error {
	runningState := int(sdktypes.SessionStateTypeRunning.ToProto())

	// Only decrement count for finished sessions with deployment
	if session.CurrentStateType > runningState && session.DeploymentID != nil {
		return tx.writer.WithContext(ctx).Model(&scheme.DeploymentSessionStats{}).
			Where("deployment_id = ? AND session_state = ?", *session.DeploymentID, session.CurrentStateType).
			Update("count", gorm.Expr("deployment_session_stats.count - 1")).Error
	}

	return nil
}

func (db *gormdb) listDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]scheme.Deployment, error) {
	q := db.listDeploymentsCommonQuery(ctx, filter)
	var ds []scheme.Deployment
	if err := q.Find(&ds).Error; err != nil {
		return nil, err
	}
	return ds, nil
}

func (db *gormdb) CreateDeployment(ctx context.Context, deployment sdktypes.Deployment) error {
	if err := deployment.Strict(); err != nil {
		return err
	}

	d := scheme.Deployment{
		Base:         based(ctx),
		ProjectID:    deployment.ProjectID().UUIDValue(),
		DeploymentID: deployment.ID().UUIDValue(),
		BuildID:      deployment.BuildID().UUIDValue(),
		State:        int32(deployment.State().ToProto()),
	}
	return translateError(db.createDeployment(ctx, &d))
}

func (db *gormdb) DeleteDeployment(ctx context.Context, deploymentID sdktypes.DeploymentID) error {
	return translateError(db.deleteDeployment(ctx, deploymentID.UUIDValue()))
}

func (db *gormdb) UpdateDeploymentState(ctx context.Context, id sdktypes.DeploymentID, state sdktypes.DeploymentState) (sdktypes.DeploymentState, error) {
	state, err := db.updateDeploymentState(ctx, id.UUIDValue(), state)
	return state, translateError(err)
}

func (db *gormdb) GetDeployment(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	d, err := db.getDeployment(ctx, id.UUIDValue())
	if d == nil || err != nil {
		return sdktypes.InvalidDeployment, translateError(err)
	}
	return scheme.ParseDeployment(*d)
}

func (db *gormdb) ListDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
	if filter.IncludeSessionStats {
		ds, err := db.listDeploymentsWithStats(ctx, filter)
		if ds == nil || err != nil {
			return nil, translateError(err)
		}
		return kittehs.TransformError(ds, scheme.ParseDeploymentWithSessionStats)
	} else {
		ds, err := db.listDeployments(ctx, filter)
		if ds == nil || err != nil {
			return nil, translateError(err)
		}
		return kittehs.TransformError(ds, scheme.ParseDeployment)
	}
}

func (db *gormdb) DeploymentHasActiveSessions(ctx context.Context, id sdktypes.DeploymentID) (bool, error) {
	r, err := db.ListSessions(ctx, sdkservices.ListSessionsFilter{
		DeploymentID: id,
		StateType:    sdktypes.SessionStateTypeCreated,
		CountOnly:    true,
	})
	if err != nil {
		return false, fmt.Errorf("count created sessions: %w", err)
	}

	if r.TotalCount > 0 {
		return true, nil
	}

	r, err = db.ListSessions(ctx, sdkservices.ListSessionsFilter{
		DeploymentID: id,
		StateType:    sdktypes.SessionStateTypeRunning,
		CountOnly:    true,
	})
	if err != nil {
		return false, fmt.Errorf("count running sessions: %w", err)
	}

	return r.TotalCount > 0, nil
}

var finalSessionStateTypes = kittehs.Transform(sdktypes.FinalSessionStateTypes, func(s sdktypes.SessionStateType) int { return s.ToInt() })

func (db *gormdb) DeactivateAllDrainedDeployments(ctx context.Context) (int, error) {
	q := db.writer.WithContext(ctx).Exec(`UPDATE deployments
SET state = ?
WHERE state = ?
AND NOT EXISTS (
    SELECT 1
    FROM sessions
    WHERE sessions.deployment_id = deployments.deployment_id
		AND sessions.current_state_type NOT IN (?)
);`,
		sdktypes.DeploymentStateInactive.ToInt(), sdktypes.DeploymentStateDraining.ToInt(), finalSessionStateTypes)

	if err := q.Error; err != nil {
		return 0, translateError(err)
	}

	return int(q.RowsAffected), nil
}

func (db *gormdb) DeactivateDrainedDeployment(ctx context.Context, did sdktypes.DeploymentID) (bool, error) {
	q := db.writer.WithContext(ctx).Exec(`UPDATE deployments
SET state = ?
WHERE state = ? AND deployment_id = ?
AND NOT EXISTS (
	SELECT 1
	FROM sessions
	WHERE sessions.deployment_id = deployments.deployment_id
		AND sessions.current_state_type NOT IN (?)
);`,
		sdktypes.DeploymentStateInactive.ToInt(), sdktypes.DeploymentStateDraining.ToInt(), did.UUIDValue(), finalSessionStateTypes)

	if err := q.Error; err != nil {
		return false, translateError(err)
	}

	return int(q.RowsAffected) > 0, nil
}
