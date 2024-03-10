package projectsgrpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	projectsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1/projectsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	MaxUploadSize int `koanf:"max_upload_size"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		MaxUploadSize: 1 * 1024 * 1024, // 1MB
	},
}

type Server struct {
	projects sdkservices.Projects

	projectsv1connect.UnimplementedProjectsServiceHandler

	cfg *Config
}

func New(cfg *Config, projects sdkservices.Projects) *Server {
	return &Server{cfg: cfg, projects: projects}
}

var _ projectsv1connect.ProjectsServiceHandler = (*Server)(nil)

func (s *Server) Init(mux *http.ServeMux) {
	path, namer := projectsv1connect.NewProjectsServiceHandler(s, connect.WithReadMaxBytes(s.cfg.MaxUploadSize))
	mux.Handle(path, namer)
}

func (s *Server) Create(ctx context.Context, req *connect.Request[projectsv1.CreateRequest]) (*connect.Response[projectsv1.CreateResponse], error) {
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

func (s *Server) Delete(ctx context.Context, req *connect.Request[projectsv1.DeleteRequest]) (*connect.Response[projectsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.projects.Delete(ctx, pid); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&projectsv1.DeleteResponse{}), nil
}

func (s *Server) Update(ctx context.Context, req *connect.Request[projectsv1.UpdateRequest]) (*connect.Response[projectsv1.UpdateResponse], error) {
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

func (s *Server) Get(ctx context.Context, req *connect.Request[projectsv1.GetRequest]) (*connect.Response[projectsv1.GetResponse], error) {
	toResponse := func(project sdktypes.Project, err error) (*connect.Response[projectsv1.GetResponse], error) {
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				return connect.NewResponse(&projectsv1.GetResponse{}), nil
			}
			return nil, sdkerrors.AsConnectError(err)
		}
		return connect.NewResponse(&projectsv1.GetResponse{Project: project.ToProto()}), nil
	}

	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if pid.IsValid() {
		return toResponse(s.projects.GetByID(ctx, pid))
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

func (s *Server) list(ctx context.Context) ([]*sdktypes.ProjectPB, error) {
	ps, err := s.projects.List(ctx)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return kittehs.Transform(ps, sdktypes.ToProto), nil
}

func (s *Server) ListForOwner(ctx context.Context, req *connect.Request[projectsv1.ListForOwnerRequest]) (*connect.Response[projectsv1.ListForOwnerResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ps, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&projectsv1.ListForOwnerResponse{Projects: ps}), nil
}

func (s *Server) List(ctx context.Context, req *connect.Request[projectsv1.ListRequest]) (*connect.Response[projectsv1.ListResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ps, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&projectsv1.ListResponse{Projects: ps}), nil
}

func (s *Server) Build(ctx context.Context, req *connect.Request[projectsv1.BuildRequest]) (*connect.Response[projectsv1.BuildResponse], error) {
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

func (s *Server) SetResources(ctx context.Context, req *connect.Request[projectsv1.SetResourcesRequest]) (*connect.Response[projectsv1.SetResourcesResponse], error) {
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

func (s *Server) DownloadResources(ctx context.Context, req *connect.Request[projectsv1.DownloadResourcesRequest]) (*connect.Response[projectsv1.DownloadResourcesResponse], error) {
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
