package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/tar"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) createProject(ctx context.Context, project *scheme.Project) error {
	return gdb.transaction(ctx, func(tx *gormdb) error {
		// ensure there is no active project with the same name (but allow deleted ones)
		var count int64
		if err := tx.wdb.
			// probably EXISTS is a bit more efficient, but it's not naturally supported by gorm
			// and we are using joins as well. First maybe a good option too, but there should be only
			// one active user project with the same name, so COUNT is also OK
			Model(&scheme.Project{}). // with model scope grom will add `deleted_at is NULL` to the query
			Where("name = ? AND org_id = ?", project.Name, project.OrgID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return gorm.ErrDuplicatedKey // active/non-deleted project was found.
		}
		return tx.wdb.Create(project).Error
	})
}

func (gdb *gormdb) deleteProject(ctx context.Context, projectID uuid.UUID) error {
	return gdb.transaction(ctx, func(tx *gormdb) error { return tx.deleteProjectAndDependents(ctx, projectID) })
}

func (gdb *gormdb) deleteProjectVars(ctx context.Context, id uuid.UUID) error {
	// NOTE: should be transactional
	db := gdb.wdb.WithContext(ctx)

	var count int64
	db.Model(&scheme.Deployment{}).Where("deleted_at is NULL and project_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("FOREIGN KEY: %w", gorm.ErrForeignKeyViolated)
	}

	if err := db.Where("project_id = ?", id).Delete(&scheme.Project{}).Error; err != nil {
		return err
	}

	return db.Where("var_id = ?", id).Delete(&scheme.Var{}).Error
}

// delete project, its envs, deployments, sessions and build.
// must be called from inside a transaction.
func (gdb *gormdb) deleteProjectAndDependents(ctx context.Context, projectID uuid.UUID) error {
	deploymentStates, err := gdb.getProjectDeploymentsStates(ctx, projectID)
	if err != nil {
		return err
	}
	var deplIDs []uuid.UUID
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
	if err = gdb.wdb.Delete(&scheme.Trigger{}, "project_id = ?", projectID).Error; err != nil {
		return err
	}

	var signalIDs []string
	if err := gdb.wdb.Model(&scheme.Signal{}).
		Joins("join connections on connections.connection_id = signals.connection_id").
		Where("connections.project_id = ?", projectID).
		Pluck("signals.signal_id", &signalIDs).Error; err != nil {
		return err
	}
	if err = gdb.wdb.Delete(&scheme.Signal{}, "signal_id IN ?", signalIDs).Error; err != nil {
		return err
	}

	// delete project connections and associated vars
	if err = gdb.deleteConnectionsAndVars(ctx, "project_id", projectID); err != nil {
		return err
	}

	if err = gdb.deleteProjectVars(ctx, projectID); err != nil {
		return err
	}

	return gdb.wdb.Delete(&scheme.Project{ProjectID: projectID}).Error
}

func (gdb *gormdb) updateProject(ctx context.Context, p *scheme.Project) error {
	return gdb.transaction(ctx, func(tx *gormdb) error {
		data := updatedBaseColumns(ctx)
		data["name"] = p.Name
		data["root_url"] = p.RootURL
		return tx.wdb.Model(&scheme.Project{ProjectID: p.ProjectID}).Updates(data).Error
	})
}

func (gdb *gormdb) getProject(ctx context.Context, projectID uuid.UUID) (*scheme.Project, error) {
	return getOne[scheme.Project](gdb.rdb.WithContext(ctx), "project_id = ?", projectID)
}

func (gdb *gormdb) getProjectByName(ctx context.Context, oid sdktypes.OrgID, projectName string) (*scheme.Project, error) {
	q := "name = ?"
	qargs := []any{projectName}

	if oid.IsValid() {
		q += " AND org_id = ?"
		qargs = append(qargs, oid.UUIDValue())
	}

	return getOne[scheme.Project](gdb.rdb.WithContext(ctx), q, qargs...)
}

func (gdb *gormdb) listProjects(ctx context.Context, oid sdktypes.OrgID) ([]scheme.Project, error) {
	q := gdb.rdb.WithContext(ctx)

	if oid.IsValid() {
		q = q.Where("org_id = ?", oid.UUIDValue())
	}

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
		Base:      based(ctx),
		OrgID:     p.OrgID().UUIDValue(),
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

func (db *gormdb) GetProjectByName(ctx context.Context, oid sdktypes.OrgID, ph sdktypes.Symbol) (sdktypes.Project, error) {
	return schemaToProject(db.getProjectByName(ctx, oid, ph.String()))
}

func (db *gormdb) ListProjects(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.Project, error) {
	ps, err := db.listProjects(ctx, oid)
	if ps == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(ps, scheme.ParseProject)
}

type DeploymentState struct {
	DeploymentID uuid.UUID
	State        int32
}

// FIXME: apply user scopes from here ---v

func (db *gormdb) getProjectDeploymentsStates(ctx context.Context, pid uuid.UUID) ([]DeploymentState, error) {
	var pds []DeploymentState
	res := db.rdb.WithContext(ctx).Model(&scheme.Deployment{}).Where("project_id = ?", pid).Find(&pds)
	return pds, res.Error
}

func (db *gormdb) GetProjectResources(ctx context.Context, pid sdktypes.ProjectID) (map[string][]byte, error) {
	res := db.rdb.WithContext(ctx).Model(&scheme.Project{}).Where("project_id = ?", pid.UUIDValue()).Select("resources").Row()
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

	res := db.wdb.WithContext(ctx).Model(&scheme.Project{}).Where("project_id = ?", pid.UUIDValue()).Update("resources", resourcesBytes)
	if res.Error != nil {
		return translateError(res.Error)
	}

	if res.RowsAffected == 0 {
		return sdkerrors.ErrNotFound
	}

	return nil
}
