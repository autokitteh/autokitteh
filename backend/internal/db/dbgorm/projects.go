package dbgorm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/backend/internal/tar"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) CreateProject(ctx context.Context, p sdktypes.Project) error {
	if !sdktypes.ProjectHasID(p) {
		db.z.DPanic("no project id supplied")
		return errors.New("project id missing")
	}

	ph := sdktypes.GetProjectName(p)

	paths, err := json.Marshal(sdktypes.GetProjectResourcePaths(p))
	if err != nil {
		return fmt.Errorf("marshal paths: %w", err)
	}

	r := scheme.Project{
		ProjectID: sdktypes.GetProjectID(p).String(),
		Name:      ph.String(),
		RootURL:   sdktypes.GetProjectResourcesRootURL(p).String(),
		Paths:     paths,
	}

	if err := db.db.WithContext(ctx).Create(&r).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) GetProjectByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	return get(db.db, ctx, scheme.ParseProject, "project_id = ?", pid.String())
}

func (db *gormdb) GetProjectByName(ctx context.Context, ph sdktypes.Name) (sdktypes.Project, error) {
	return get(db.db, ctx, scheme.ParseProject, "name = ?", ph.String())
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

// No delete project ?

func (db *gormdb) GetProjectResources(ctx context.Context, pid sdktypes.ProjectID) (map[string][]byte, error) {
	res := db.db.WithContext(ctx).Model(&scheme.Project{}).Where("project_id = ?", pid.String()).Select("resources").Row()
	var resources []byte
	if err := res.Scan(&resources); err != nil {
		return nil, translateError(err)
	}

	if len(resources) == 0 {
		return map[string][]byte{}, nil
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
