package projectsauth

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
)

type projects struct {
	projects sdkservices.Projects
}

func Wrap(in sdkservices.Projects) sdkservices.Projects { return &projects{projects: in} }

func (o *projects) Create(ctx context.Context, project sdktypes.Project) (sdktypes.ProjectID, error) {
	if userID := authcontext.GetAuthnUserID(ctx); !project.OwnerID().IsValid() && userID.IsValid() {
		project = project.WithOwnerID(sdktypes.NewOwnerID(userID))
	}

	if !project.OwnerID().IsValid() {
		return sdktypes.InvalidProjectID, sdkerrors.NewInvalidArgumentError("owner: missing")
	}

	return o.projects.Create(ctx, project)
}

func (o *projects) Delete(ctx context.Context, pid sdktypes.ProjectID) error {
	return o.projects.Delete(ctx, pid)
}

func (o *projects) GetByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	// TODO: Make sure somone can't get a project they don't own or member of its org.
	return o.projects.GetByID(ctx, pid)
}

func (o *projects) GetByName(ctx context.Context, owid sdktypes.OwnerID, n sdktypes.Symbol) (sdktypes.Project, error) {
	if userID := authcontext.GetAuthnUserID(ctx); !owid.IsValid() && userID.IsValid() {
		owid = sdktypes.NewOwnerID(userID)
	}

	return o.projects.GetByName(ctx, owid, n)
}

func (o *projects) ListForOwner(ctx context.Context, owid sdktypes.OwnerID) ([]sdktypes.Project, error) {
	if userID := authcontext.GetAuthnUserID(ctx); !owid.IsValid() && userID.IsValid() {
		owid = sdktypes.NewOwnerID(userID)
	}

	return o.projects.ListForOwner(ctx, owid)
}

func (o *projects) Build(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.BuildID, error) {
	return o.projects.Build(ctx, projectID)
}

func (o *projects) SetResources(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) error {
	return o.projects.SetResources(ctx, projectID, resources)
}

func (o *projects) DownloadResources(ctx context.Context, projectID sdktypes.ProjectID) (map[string][]byte, error) {
	return o.projects.DownloadResources(ctx, projectID)
}
