package projects

import (
	"bytes"
	"context"
	"fmt"
	"os"

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

func (o *Projects) Create(ctx context.Context, project sdktypes.Project) (sdktypes.ProjectID, error) {
	project, err := project.Update(func(pb *sdktypes.ProjectPB) {
		pb.ProjectId = sdktypes.NewProjectID().String()
	})
	if err != nil {
		return nil, err
	}

	if project, err = sdktypes.ToStrictProject(project); err != nil {
		return nil, err
	}

	if err := o.DB.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	return sdktypes.GetProjectID(project), nil
}

func (o *Projects) GetByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	// TODO: Make sure somone can't get a project they don't own or member of its org.
	return sdkerrors.IgnoreNotFoundErr(o.DB.GetProjectByID(ctx, pid))
}

func (o *Projects) GetByName(ctx context.Context, n sdktypes.Name) (sdktypes.Project, error) {
	// TODO: Make sure somone can't get a project they don't own or member of its org.
	return sdkerrors.IgnoreNotFoundErr(o.DB.GetProjectByName(ctx, n))
}

func (o *Projects) List(ctx context.Context) ([]sdktypes.Project, error) {
	return o.DB.ListProjects(ctx)
}

func (o *Projects) Build(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.BuildID, error) {
	p, err := o.DB.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, nil
	}

	url := sdktypes.GetProjectResourcesRootURL(p)
	if url.Scheme != "" && url.Scheme != "file" {
		return nil, fmt.Errorf("%w: only file resources are supported", sdkerrors.ErrNotImplemented)
	}

	bi, err := sdkbuild.Build(
		ctx,
		o.Runtimes,
		os.DirFS(url.Path),
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

	return o.Builds.Save(ctx, build, buf.Bytes())
}

func (o *Projects) SetResources(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) error {
	return o.DB.SetProjectResources(ctx, projectID, resources)
}

func (o *Projects) DownloadResources(ctx context.Context, projectID sdktypes.ProjectID) (map[string][]byte, error) {
	return o.DB.GetProjectResources(ctx, projectID)
}
