package projectsstoregrpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbprojectsvc "go.autokitteh.dev/idl/go/projectsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
	"go.autokitteh.dev/sdk/api/apiaccount"
	"go.autokitteh.dev/sdk/api/apiproject"
)

type Store struct{ Client pbprojectsvc.ProjectsClient }

var _ projectsstore.Store = &Store{}

func (as *Store) Create(ctx context.Context, aname apiaccount.AccountName, pid apiproject.ProjectID, d *apiproject.ProjectSettings) (apiproject.ProjectID, error) {
	resp, err := as.Client.CreateProject(
		ctx,
		&pbprojectsvc.CreateProjectRequest{Settings: d.PB(), Id: pid.String(), AccountName: aname.String()},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				return "", projectsstore.ErrAlreadyExists
			} else if e.Code() == codes.FailedPrecondition {
				return "", projectsstore.ErrInvalidAccount
			}
		}

		return "", fmt.Errorf("create: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return "", fmt.Errorf("resp validate: %w", err)
	}

	return apiproject.ProjectID(resp.Id), nil
}

func (as *Store) Update(
	ctx context.Context,
	id apiproject.ProjectID,
	d *apiproject.ProjectSettings,
) error {
	resp, err := as.Client.UpdateProject(
		ctx,
		&pbprojectsvc.UpdateProjectRequest{
			Id:       id.String(),
			Settings: d.PB(),
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return projectsstore.ErrNotFound
			}
		}

		return fmt.Errorf("update: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return fmt.Errorf("resp validate: %w", err)
	}

	return nil
}

func (as *Store) Get(ctx context.Context, id apiproject.ProjectID) (*apiproject.Project, error) {
	resp, err := as.Client.GetProject(
		ctx,
		&pbprojectsvc.GetProjectRequest{Id: id.String()},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, projectsstore.ErrNotFound
			}
		}

		return nil, fmt.Errorf("get: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	a, err := apiproject.ProjectFromProto(resp.Project)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (db *Store) BatchGet(ctx context.Context, ids []apiproject.ProjectID) (map[apiproject.ProjectID]*apiproject.Project, error) {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}

	resp, err := db.Client.GetProjects(
		ctx,
		&pbprojectsvc.GetProjectsRequest{Ids: strs},
	)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	m := make(map[apiproject.ProjectID]*apiproject.Project, len(ids))
	for _, id := range ids {
		m[id] = nil
	}
	for _, pbp := range resp.Projects {
		if m[apiproject.ProjectID(pbp.Id)], err = apiproject.ProjectFromProto(pbp); err != nil {
			return nil, fmt.Errorf("invalid project %q: %w", pbp.Id, err)
		}
	}

	return m, nil
}

func (as *Store) Setup(ctx context.Context) error    { return fmt.Errorf("not supported through grpc") }
func (as *Store) Teardown(ctx context.Context) error { return fmt.Errorf("not supported through grpc") }
