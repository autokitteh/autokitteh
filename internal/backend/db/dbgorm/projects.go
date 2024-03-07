package dbgorm

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/tar"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) createProject(ctx context.Context, p scheme.Project) error {
	return db.db.WithContext(ctx).Create(&p).Error
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
	return translateError(db.createProject(ctx, project))
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

func (db *gormdb) GetProjectByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseProject, "project_id = ?", pid.String())
}

func (db *gormdb) GetProjectByName(ctx context.Context, ph sdktypes.Symbol) (sdktypes.Project, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseProject, "name = ?", ph.String())
}

func (db *gormdb) ListProjects(ctx context.Context) ([]sdktypes.Project, error) {
	var rs []scheme.Project
	q := db.db.WithContext(ctx)

	err := q.Order("project_id").Find(&rs).Error
	if err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(rs, scheme.ParseProject)
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
