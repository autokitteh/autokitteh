package projectsgrpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	projectsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1/projectsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	projects sdkservices.Projects

	projectsv1connect.UnimplementedProjectsServiceHandler
}

var _ projectsv1connect.ProjectsServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, projects sdkservices.Projects) {
	srv := server{projects: projects}

	path, namer := projectsv1connect.NewProjectsServiceHandler(&srv)
	mux.Handle(path, namer)
}

func (s *server) Create(ctx context.Context, req *connect.Request[projectsv1.CreateRequest]) (*connect.Response[projectsv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	project, err := sdktypes.ProjectFromProto(msg.Project)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := s.projects.Create(ctx, project)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&projectsv1.CreateResponse{ProjectId: uid.String()}), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[projectsv1.UpdateRequest]) (*connect.Response[projectsv1.UpdateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	project, err := sdktypes.ProjectFromProto(msg.Project)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.projects.Update(ctx, project); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&projectsv1.UpdateResponse{}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[projectsv1.GetRequest]) (*connect.Response[projectsv1.GetResponse], error) {
	toResponse := func(project sdktypes.Project, err error) (*connect.Response[projectsv1.GetResponse], error) {
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				return connect.NewResponse(&projectsv1.GetResponse{}), nil
			} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}

			return nil, connect.NewError(connect.CodeUnknown, err)
		}

		return connect.NewResponse(&projectsv1.GetResponse{Project: project.ToProto()}), nil
	}

	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if uid.IsValid() {
		return toResponse(s.projects.GetByID(ctx, uid))
	}

	n, err := sdktypes.StrictParseSymbol(msg.Name)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if !n.IsValid() {
		// essentially should never happen since we validate existance of name xor uid
		// in proto. Hence Unknown error.
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("missing name"))
	}

	return toResponse(s.projects.GetByName(ctx, n))
}

func (s *server) list(ctx context.Context) ([]*sdktypes.ProjectPB, error) {
	ps, err := s.projects.List(ctx)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return kittehs.Transform(ps, sdktypes.ToProto), nil
}

func (s *server) ListForOwner(ctx context.Context, req *connect.Request[projectsv1.ListForOwnerRequest]) (*connect.Response[projectsv1.ListForOwnerResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ps, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&projectsv1.ListForOwnerResponse{Projects: ps}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[projectsv1.ListRequest]) (*connect.Response[projectsv1.ListResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ps, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&projectsv1.ListResponse{Projects: ps}), nil
}

func (s *server) Build(ctx context.Context, req *connect.Request[projectsv1.BuildRequest]) (*connect.Response[projectsv1.BuildResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if !pid.IsValid() {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project_id: %w", err))
	}

	bid, err := s.projects.Build(ctx, pid)
	if err != nil {
		if err, ok := sdktypes.FromError(err); ok {
			return connect.NewResponse(&projectsv1.BuildResponse{Error: err.ToProto()}), nil
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&projectsv1.BuildResponse{BuildId: bid.String()}), nil
}

func (s *server) SetResources(ctx context.Context, req *connect.Request[projectsv1.SetResourcesRequest]) (*connect.Response[projectsv1.SetResourcesResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if !pid.IsValid() {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project_id: %w", err))
	}

	if err := s.projects.SetResources(ctx, pid, msg.Resources); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&projectsv1.SetResourcesResponse{}), nil
}

func (s *server) DownloadResources(ctx context.Context, req *connect.Request[projectsv1.DownloadResourcesRequest]) (*connect.Response[projectsv1.DownloadResourcesResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if !pid.IsValid() {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("project_id: %w", err))
	}

	resources, err := s.projects.DownloadResources(ctx, pid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&projectsv1.DownloadResourcesResponse{Resources: resources}), nil
}
