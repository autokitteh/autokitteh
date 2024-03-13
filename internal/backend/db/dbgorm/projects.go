package dbgorm

import (
	"context"
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/tar"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) createProject(ctx context.Context, p *scheme.Project) error {
	return db.db.WithContext(ctx).Create(p).Error
}

func (db *gormdb) CreateProject(ctx context.Context, p sdktypes.Project) error {
	if !p.ID().IsValid() {
		db.z.DPanic("no project id supplied")
		return errors.New("project id missing")
	}

	project := scheme.Project{
		ProjectID: p.ID().String(),
		Name:      p.Name().String(),
	}
	return translateError(db.createProject(ctx, &project))
}

func (db *gormdb) deleteProject(ctx context.Context, projectID string) error {
	return db.deleteProjectAndDependents(ctx, projectID)
}

// delete project, its envs, deployments, sessions and build
func (db *gormdb) deleteProjectAndDependents(ctx context.Context, projectID string) error {
	// NOTE: should be transactional

	deploymentStates, err := db.getProjectDeployments(ctx, projectID)
	if err != nil {
		return err
	}
	var deplIDs []string
	for _, de := range deploymentStates {
		if de.State != int32(sdktypes.DeploymentStateInactive.ToProto()) {
			return fmt.Errorf("%w: project <%s>: cannot delete non-inactive deployment <%s> in state <%s>",
				sdkerrors.ErrFailedPrecondition, projectID, de.DeploymentID,
				deploymentsv1.DeploymentState(de.State).String())
		}
		deplIDs = append(deplIDs, de.DeploymentID)
	}

	// TODO: consider deleting deployments from envs
	if err = db.deleteDeploymentsAndDependents(ctx, deplIDs); err != nil {
		return err
	}

	gormDB := db.db.WithContext(ctx)

	if err = gormDB.Delete(&scheme.Trigger{}, "project_id = ?", projectID).Error; err != nil {
		return err
	}

	if err = gormDB.Delete(&scheme.Connection{}, "project_id = ?", projectID).Error; err != nil {
		return err
	}

	envIDs, err := db.getProjectEnvs(ctx, projectID)
	if err != nil {
		return err
	}
	if err = db.deleteEnvs(ctx, envIDs); err != nil {
		return err
	}

	return gormDB.Delete(&scheme.Project{ProjectID: projectID}).Error
}

func (db *gormdb) DeleteProject(ctx context.Context, projectID sdktypes.ProjectID) error {
	return db.transaction(ctx, func(tx *tx) error {
		return translateError(tx.deleteProject(ctx, projectID.String()))
	})
}

func (db *gormdb) UpdateProject(ctx context.Context, p sdktypes.Project) error {
	r := scheme.Project{
		ProjectID: p.ID().String(),
		Name:      p.Name().String(),
	}

	if err := db.db.WithContext(ctx).Updates(&r).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) getProject(ctx context.Context, projectID string) (*scheme.Project, error) {
	return getOne(db.db, ctx, scheme.Project{}, "project_id = ?", projectID)
}

func (db *gormdb) getProjectByName(ctx context.Context, projectName string) (*scheme.Project, error) {
	return getOne(db.db, ctx, scheme.Project{}, "name = ?", projectName)
}

func schemaToSDKProject(p *scheme.Project, err error) (sdktypes.Project, error) {
	if p == nil || err != nil {
		return sdktypes.InvalidProject, translateError(err)
	}
	return scheme.ParseProject(*p)
}

func (db *gormdb) GetProjectByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	return schemaToSDKProject(db.getProject(ctx, pid.String()))
}

func (db *gormdb) GetProjectByName(ctx context.Context, ph sdktypes.Symbol) (sdktypes.Project, error) {
	return schemaToSDKProject(db.getProjectByName(ctx, ph.String()))
}

type DeploymentState struct {
	DeploymentID string
	State        int32
}

func (db *gormdb) getProjectDeployments(ctx context.Context, pid string) ([]DeploymentState, error) {
	var pds []DeploymentState
	res := db.db.WithContext(ctx).Model(&scheme.Deployment{}).
		Joins("join Envs on Envs.env_id = Deployments.env_id").
		Where("Envs.project_id = ?", pid).
		Select("DISTINCT Deployments.deployment_id, Deployments.state").
		Find(&pds)
	return pds, res.Error
}

func (db *gormdb) getProjectEnvs(ctx context.Context, pid string) (envIDs []string, err error) {
	err = db.db.WithContext(ctx).Model(&scheme.Env{}).Where("project_id = ?", pid).Pluck("env_id", &envIDs).Error
	return envIDs, err
}

func (db *gormdb) listProjects(ctx context.Context) ([]scheme.Project, error) {
	q := db.db.WithContext(ctx).Order("project_id")

	var ps []scheme.Project
	if err := q.Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (db *gormdb) ListProjects(ctx context.Context) ([]sdktypes.Project, error) {
	ps, err := db.listProjects(ctx)
	if ps == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(ps, scheme.ParseProject)
}

func (db *gormdb) GetProjectResources(ctx context.Context, pid sdktypes.ProjectID) (map[string][]byte, error) {
	res := db.db.WithContext(ctx).Model(&scheme.Project{}).Where("project_id = ?", pid.String()).Select("resources").Row()
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

	res := db.db.WithContext(ctx).Model(&scheme.Project{}).Where("project_id = ?", pid.String()).Update("resources", resourcesBytes)
	if res.Error != nil {
		return translateError(res.Error)
	}

	if res.RowsAffected == 0 {
		return sdkerrors.ErrNotFound
	}

	return nil
}
