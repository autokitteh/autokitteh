package dbgorm

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/tar"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"gorm.io/gorm"
)

func (gdb *gormdb) withUserProjects(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "project")
}

func (gdb *gormdb) createProject(ctx context.Context, p *scheme.Project) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		// ensure there is no active project with the same name (but allow deleted ones)
		// - `deleted_at is NULL` will be added automatically to the query scope
		// - we use Find, since it won't return and report/log ErrRecordNotFound
		var count int64
		if err := tx.withUserProjects(ctx).Model(&scheme.Project{}).
			Where("name = ?", p.Name).Limit(1).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return gorm.ErrDuplicatedKey // active/non-deleted project was found.
		}
		return createEntityWithOwnership(ctx, tx.gormdb.db, p)
	})
}

func (gdb *gormdb) deleteProject(ctx context.Context, projectID sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isUserEntity(ctx, projectID); err != nil {
			return err
		}
		return tx.deleteProjectAndDependents(ctx, projectID)
	})
}

// delete project, its envs, deployments, sessions and build
func (gdb *gormdb) deleteProjectAndDependents(ctx context.Context, projectID sdktypes.UUID) error {
	// NOTE: should be transactional

	db := gdb.db.WithContext(ctx)

	deploymentStates, err := gdb.getProjectDeployments(ctx, projectID)
	if err != nil {
		return err
	}
	var deplIDs []sdktypes.UUID
	for _, de := range deploymentStates {
		if de.State != int32(sdktypes.DeploymentStateInactive.ToProto()) {
			return fmt.Errorf("%w: project <%s>: cannot delete non-inactive deployment <%s> in state <%s>",
				sdkerrors.ErrFailedPrecondition, projectID, de.DeploymentID,
				deploymentsv1.DeploymentState(de.State).String())
		}
		deplIDs = append(deplIDs, de.DeploymentID)
	}

	// TODO: consider deleting deployments from envs
	if err = gdb.deleteDeploymentsAndDependents(ctx, deplIDs); err != nil {
		return err
	}

	// Connection is referenced by signals and triggers, so delete them first.
	// NOTE that signals, triggers and connections are hard-deleted now
	if err = db.Delete(&scheme.Trigger{}, "project_id = ?", projectID).Error; err != nil {
		return err
	}

	var signalIDs []string
	if err := db.Model(&scheme.Signal{}).
		Joins("join connections on connections.connection_id = signals.connection_id").
		Where("connections.project_id = ?", projectID).
		Pluck("signals.signal_id", &signalIDs).Error; err != nil {
		return err
	}
	if err = db.Delete(&scheme.Signal{}, "signal_id IN ?", signalIDs).Error; err != nil {
		return err
	}

	if err = db.Delete(&scheme.Connection{}, "project_id = ?", projectID).Error; err != nil {
		return err
	}

	envIDs, err := gdb.getProjectEnvs(ctx, projectID)
	if err != nil {
		return err
	}
	if err = gdb.deleteEnvs(ctx, envIDs); err != nil {
		return err
	}

	return db.Delete(&scheme.Project{ProjectID: projectID}).Error
}

func (gdb *gormdb) updateProject(ctx context.Context, p *scheme.Project) error {
	// REVIEW: security? any specific fields to allow? resources to disallow?
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isUserEntity(ctx, p.ProjectID); err != nil {
			return err
		}
		return tx.db.Updates(p).Error
	})
}

func (gdb *gormdb) getProject(ctx context.Context, projectID sdktypes.UUID) (*scheme.Project, error) {
	return getOne[scheme.Project](gdb.withUserProjects(ctx), "project_id = ?", projectID)
}

func (gdb *gormdb) getProjectByName(ctx context.Context, projectName string) (*scheme.Project, error) {
	return getOne[scheme.Project](gdb.withUserProjects(ctx), "name = ?", projectName)
}

func (gdb *gormdb) listProjects(ctx context.Context) ([]scheme.Project, error) {
	q := gdb.withUserProjects(ctx)

	var ps []scheme.Project
	if err := q.Find(&ps).Order("project_id").Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (db *gormdb) CreateProject(ctx context.Context, p sdktypes.Project) error {
	if err := p.Strict(); err != nil {
		return err
	}

	project := scheme.Project{
		ProjectID: p.ID().UUIDValue(),
		Name:      p.Name().String(),
	}
	return translateError(db.createProject(ctx, &project))
}

func (gdb *gormdb) DeleteProject(ctx context.Context, projectID sdktypes.ProjectID) error {
	return translateError(gdb.deleteProjectAndDependents(ctx, projectID.UUIDValue()))
}

func (gdb *gormdb) UpdateProject(ctx context.Context, project sdktypes.Project) error {
	p := scheme.Project{
		ProjectID: project.ID().UUIDValue(),
		Name:      project.Name().String(),
	}
	return translateError(gdb.updateProject(ctx, &p))
}

func schemaToProject(p *scheme.Project, err error) (sdktypes.Project, error) {
	if p == nil || err != nil {
		return sdktypes.InvalidProject, translateError(err)
	}
	return scheme.ParseProject(*p)
}

func (db *gormdb) GetProjectByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	return schemaToProject(db.getProject(ctx, pid.UUIDValue()))
}

func (db *gormdb) GetProjectByName(ctx context.Context, ph sdktypes.Symbol) (sdktypes.Project, error) {
	return schemaToProject(db.getProjectByName(ctx, ph.String()))
}

func (db *gormdb) ListProjects(ctx context.Context) ([]sdktypes.Project, error) {
	ps, err := db.listProjects(ctx)
	if ps == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(ps, scheme.ParseProject)
}

type DeploymentState struct {
	DeploymentID sdktypes.UUID
	State        int32
}

// FIXME: apply user scopes from here ---v

func (db *gormdb) getProjectDeployments(ctx context.Context, pid sdktypes.UUID) ([]DeploymentState, error) {
	var pds []DeploymentState
	res := db.db.WithContext(ctx).Model(&scheme.Deployment{}).
		Joins("join Envs on Envs.env_id = Deployments.env_id").
		Where("Envs.project_id = ?", pid).
		Select("DISTINCT Deployments.deployment_id, Deployments.state").
		Find(&pds)
	return pds, res.Error
}

func (db *gormdb) getProjectEnvs(ctx context.Context, pid sdktypes.UUID) (envIDs []sdktypes.UUID, err error) {
	err = db.db.WithContext(ctx).Model(&scheme.Env{}).Where("project_id = ?", pid).Pluck("env_id", &envIDs).Error
	return envIDs, err
}

func (db *gormdb) GetProjectResources(ctx context.Context, pid sdktypes.ProjectID) (map[string][]byte, error) {
	res := db.db.WithContext(ctx).Model(&scheme.Project{}).Where("project_id = ?", pid.UUIDValue()).Select("resources").Row()
	var resources []byte
	if err := res.Scan(&resources); err != nil {
		return nil, translateError(err)
	}

	if len(resources) == 0 {
		return nil, nil
	}

	t, err := tar.FromBytes(resources, true)
	if err != nil {
		return nil, err
	}

	return t.Content()
}

func (db *gormdb) SetProjectResources(ctx context.Context, pid sdktypes.ProjectID, resources map[string][]byte) error {
	t := tar.NewTarFile()
	for name, data := range resources {
		t.Add(name, data)
	}
	resourcesBytes, err := t.Bytes(true)
	if err != nil {
		return err
	}

	res := db.db.WithContext(ctx).Model(&scheme.Project{}).Where("project_id = ?", pid.UUIDValue()).Update("resources", resourcesBytes)
	if res.Error != nil {
		return translateError(res.Error)
	}

	if res.RowsAffected == 0 {
		return sdkerrors.ErrNotFound
	}

	return nil
}
