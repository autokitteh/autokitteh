package projectsstoregrpcsvc

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbproject "go.autokitteh.dev/idl/go/project"
	pbprojectsvc "go.autokitteh.dev/idl/go/projectsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
	"go.autokitteh.dev/sdk/api/apiaccount"
	"go.autokitteh.dev/sdk/api/apiproject"
	"github.com/autokitteh/L"
)

type Svc struct {
	pbprojectsvc.UnimplementedProjectsServer

	Store projectsstore.Store

	L L.Nullable
}

var _ pbprojectsvc.ProjectsServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pbprojectsvc.RegisterProjectsServer(srv, s)

	if gw != nil {
		if err := pbprojectsvc.RegisterProjectsHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) CreateProject(ctx context.Context, req *pbprojectsvc.CreateProjectRequest) (*pbprojectsvc.CreateProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apiproject.ProjectSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	id := apiproject.ProjectID(req.Id)
	if id == "" {
		id = projectsstore.AutoProjectID
	}

	if id, err = s.Store.Create(ctx, apiaccount.AccountName(req.AccountName), id, d); err != nil {
		if errors.Is(err, projectsstore.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "create: %v", err)
		} else if errors.Is(err, projectsstore.ErrInvalidAccount) {
			return nil, status.Errorf(codes.FailedPrecondition, "create: %v", err)
		}

		return nil, status.Errorf(codes.Unknown, "create: %v", err)
	}

	return &pbprojectsvc.CreateProjectResponse{
		Id: id.String(),
	}, nil
}

func (s *Svc) UpdateProject(ctx context.Context, req *pbprojectsvc.UpdateProjectRequest) (*pbprojectsvc.UpdateProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apiproject.ProjectSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	id := apiproject.ProjectID(req.Id)

	if err := s.Store.Update(ctx, id, d); err != nil {
		return nil, status.Errorf(codes.Unknown, "update: %v", err)
	}

	return &pbprojectsvc.UpdateProjectResponse{}, nil
}

func (s *Svc) GetProject(ctx context.Context, req *pbprojectsvc.GetProjectRequest) (*pbprojectsvc.GetProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	id := apiproject.ProjectID(req.Id)

	a, err := s.Store.Get(ctx, id)
	if err != nil {
		if errors.Is(err, projectsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}
	return &pbprojectsvc.GetProjectResponse{Project: a.PB()}, nil
}

func (s *Svc) GetProjects(ctx context.Context, req *pbprojectsvc.GetProjectsRequest) (*pbprojectsvc.GetProjectsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	ids := make([]apiproject.ProjectID, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = apiproject.ProjectID(id)
	}

	ps, err := s.Store.BatchGet(ctx, ids)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	pbps := make([]*pbproject.Project, 0, len(ps))
	for _, v := range ps {
		if v != nil {
			pbps = append(pbps, v.PB())
		}
	}

	return &pbprojectsvc.GetProjectsResponse{Projects: pbps}, nil
}
