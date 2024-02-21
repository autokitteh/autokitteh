package projects

import (
	"bytes"
	"context"
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/db"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuild"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Projects struct {
	fx.In

	Z        *zap.Logger
	DB       db.DB
	Builds   sdkservices.Builds
	Runtimes sdkservices.Runtimes
}

func New(p Projects) sdkservices.Projects { return &p }

func (ps *Projects) Create(ctx context.Context, project sdktypes.Project) (sdktypes.ProjectID, error) {
	project, err := project.Update(func(pb *sdktypes.ProjectPB) {
		pb.ProjectId = sdktypes.NewProjectID().String()
	})
	if err != nil {
		return nil, err
	}

	if project, err = sdktypes.ToStrictProject(project); err != nil {
		return nil, err
	}

	if err := ps.DB.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	return sdktypes.GetProjectID(project), nil
}

func (ps *Projects) Update(ctx context.Context, project sdktypes.Project) error {
	return ps.DB.UpdateProject(ctx, project)
}

func (ps *Projects) GetByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	// TODO: Make sure somone can't get a project they don't own or member of its org.
	return sdkerrors.IgnoreNotFoundErr(ps.DB.GetProjectByID(ctx, pid))
}

func (ps *Projects) GetByName(ctx context.Context, n sdktypes.Name) (sdktypes.Project, error) {
	// TODO: Make sure somone can't get a project they don't own or member of its org.
	return sdkerrors.IgnoreNotFoundErr(ps.DB.GetProjectByName(ctx, n))
}

func (ps *Projects) List(ctx context.Context) ([]sdktypes.Project, error) {
	return ps.DB.ListProjects(ctx)
}

func (ps *Projects) Build(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.BuildID, error) {
	p, err := ps.DB.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, nil
	}

	fs, err := ps.openProjectResourcesFS(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if fs == nil {
		return nil, errors.New("no resources set")
	}

	bi, err := sdkbuild.Build(
		ctx,
		ps.Runtimes,
		fs,
		sdktypes.GetProjectResourcePaths(p),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := bi.Write(&buf); err != nil {
		return nil, err
	}

	build := kittehs.Must1(sdktypes.BuildFromProto(&sdktypes.BuildPB{
		ProjectId: projectID.String(),
	}))

	return ps.Builds.Save(ctx, build, buf.Bytes())
}

func (ps *Projects) SetResources(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) error {
	return ps.DB.SetProjectResources(ctx, projectID, resources)
}

func (ps *Projects) DownloadResources(ctx context.Context, projectID sdktypes.ProjectID) (map[string][]byte, error) {
	return ps.DB.GetProjectResources(ctx, projectID)
}
