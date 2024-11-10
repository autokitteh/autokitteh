package dbgorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/tar"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) withUserProjects(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "project")
}

func (gdb *gormdb) createProject(ctx context.Context, project *scheme.Project) error {
	createFunc := func(tx *gorm.DB, uid string) error {
		// ensure there is no active project with the same name (but allow deleted ones)
		var count int64
		if err := tx.
			// probably EXISTS is a bit more efficient, but it's not naturally supported by gorm
			// and we are using joins as well. First maybe a good option too, but there should be only
			// one active user project with the same name, so COUNT is also OK
			Model(&scheme.Project{}). // with model scope grom will add `deleted_at is NULL` to the query
			Scopes(withUserEntity(ctx, gdb, "project", uid)).
			Where("name = ?", project.Name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return gorm.ErrDuplicatedKey // active/non-deleted project was found.
		}
		return tx.Create(project).Error
	}
	return gdb.createEntityWithOwnership(ctx, createFunc, project)
}

func (gdb *gormdb) deleteProject(ctx context.Context, projectID sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(tx.ctx, projectID); err != nil {
			return err
		}
		return tx.deleteProjectAndDependents(tx.ctx, projectID)
	})
}

func (gdb *gormdb) deleteProjectVars(ctx context.Context, id sdktypes.UUID) error {
	// NOTE: should be transactional
	db := gdb.db.WithContext(ctx)

	// enforce foreign keys constrains while soft-deleting - should be no active deployments
	var count int64
	db.Model(&scheme.Deployment{}).Where("deleted_at is NULL and project_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("FOREIGN KEY: %w", gorm.ErrForeignKeyViolated)
	}

	// REVIEW: should we allow deleting only user envs and skipping the others.
	// The check below will fail of any of provided envs cannot be deleted
	if err := gdb.isCtxUserEntity(ctx, id); err != nil {
		return err
	}

	// delete envs and associated vars
	if err := db.Where("project_id = ?", id).Delete(&scheme.Project{}).Error; err != nil {
		return err
	}

	return db.Where("var_id = ?", id).Delete(&scheme.Var{}).Error
}

// delete project, its envs, deployments, sessions and build.
// must be called from inside a transaction.
func (gdb *gormdb) deleteProjectAndDependents(ctx context.Context, projectID sdktypes.UUID) error {
	deploymentStates, err := gdb.getProjectDeploymentsStates(ctx, projectID)
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
	if err = gdb.db.Delete(&scheme.Trigger{}, "project_id = ?", projectID).Error; err != nil {
		return err
	}

	var signalIDs []string
	if err := gdb.db.Model(&scheme.Signal{}).
		Joins("join connections on connections.connection_id = signals.connection_id").
		Where("connections.project_id = ?", projectID).
		Pluck("signals.signal_id", &signalIDs).Error; err != nil {
		return err
	}
	if err = gdb.db.Delete(&scheme.Signal{}, "signal_id IN ?", signalIDs).Error; err != nil {
		return err
	}

	// delete project connections and associated vars
	if err = gdb.deleteConnectionsAndVars("project_id", projectID); err != nil {
		return err
	}

	if err = gdb.deleteProjectVars(ctx, projectID); err != nil {
		return err
	}

	return gdb.db.Delete(&scheme.Project{ProjectID: projectID}).Error
}

func (gdb *gormdb) updateProject(ctx context.Context, p *scheme.Project) error {
	// REVIEW: security? any specific fields to allow? resources to disallow?
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(tx.ctx, p.ProjectID); err != nil {
			return err
		}

		data := map[string]any{"Name": p.Name, "RootURL": p.RootURL}
		allowedFields := []string{"Name", "RootURL"} // NOTE: resources are updated via SetResurces
		return tx.db.Model(&scheme.Project{ProjectID: p.ProjectID}).Select(allowedFields).Updates(data).Error
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
	return translateError(gdb.deleteProject(ctx, projectID.UUIDValue()))
}

func (gdb *gormdb) UpdateProject(ctx context.Context, p sdktypes.Project) error {
	if err := p.Strict(); err != nil {
		return err
	}

	project := scheme.Project{
		ProjectID: p.ID().UUIDValue(),
		Name:      p.Name().String(),
	}

	return translateError(gdb.updateProject(ctx, &project))
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

func (db *gormdb) getProjectDeploymentsStates(ctx context.Context, pid sdktypes.UUID) ([]DeploymentState, error) {
	var pds []DeploymentState
	res := db.db.WithContext(ctx).Model(&scheme.Deployment{}).Where("project_id = ?", pid).Find(&pds)
	return pds, res.Error
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
