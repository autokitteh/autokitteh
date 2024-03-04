package projects

import (
	"bytes"
	"context"
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
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
	project = project.WithNewID()

	if !project.Name().IsValid() {
		project = project.WithName(sdktypes.NewRandomSymbol())
	}

	if err := project.Strict(); err != nil {
		return sdktypes.InvalidProjectID, err
	}

	if err := ps.DB.CreateProject(ctx, project); err != nil {
		return sdktypes.InvalidProjectID, err
	}

	return project.ID(), nil
}

func (ps *Projects) Delete(ctx context.Context, pid sdktypes.ProjectID) error {
	// TODO: Make sure somone can't delete a project they don't own or member of its org.
	return nil
}

func (ps *Projects) Update(ctx context.Context, project sdktypes.Project) error {
	return ps.DB.UpdateProject(ctx, project)
}

func (ps *Projects) GetByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	// TODO: Make sure somone can't get a project they don't own or member of its org.
	return sdkerrors.IgnoreNotFoundErr(ps.DB.GetProjectByID(ctx, pid))
}

func (ps *Projects) GetByName(ctx context.Context, n sdktypes.Symbol) (sdktypes.Project, error) {
	// TODO: Make sure somone can't get a project they don't own or member of its org.
	return sdkerrors.IgnoreNotFoundErr(ps.DB.GetProjectByName(ctx, n))
}

func (ps *Projects) List(ctx context.Context) ([]sdktypes.Project, error) {
	return ps.DB.ListProjects(ctx)
}

func (ps *Projects) Build(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.BuildID, error) {
	fs, err := ps.openProjectResourcesFS(ctx, projectID)
	if err != nil {
		return sdktypes.InvalidBuildID, err
	}

	if fs == nil {
		return sdktypes.InvalidBuildID, errors.New("no resources set")
	}

	bi, err := sdkruntimes.Build(
		ctx,
		ps.Runtimes,
		fs,
		nil,
		nil,
	)
	if err != nil {
		return sdktypes.InvalidBuildID, err
	}

	var buf bytes.Buffer

	if err := bi.Write(&buf); err != nil {
		return sdktypes.InvalidBuildID, err
	}

	return ps.Builds.Save(ctx, sdktypes.NewBuild(), buf.Bytes())
}

func (ps *Projects) SetResources(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) error {
	return ps.DB.SetProjectResources(ctx, projectID, resources)
}

func (ps *Projects) DownloadResources(ctx context.Context, projectID sdktypes.ProjectID) (map[string][]byte, error) {
	return ps.DB.GetProjectResources(ctx, projectID)
}
