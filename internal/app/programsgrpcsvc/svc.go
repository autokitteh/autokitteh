package programsstoregrpcsvc

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.autokitteh.dev/idl/go/programssvc"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/programs"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"
)

type Svc struct {
	pb.UnimplementedProgramsServer

	Programs *programs.Programs

	L L.Nullable
}

var _ pb.ProgramsServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pb.RegisterProgramsServer(srv, s)

	if gw != nil {
		if err := pb.RegisterProgramsHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	err := req.Validate()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	var path *apiprogram.Path

	if req.RawPath != "" {
		path, err = apiprogram.ParsePathString(req.RawPath)
	} else {
		path, err = apiprogram.PathFromProto(req.Path)
	}

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid path: %v", err)
	}

	v, err := s.Programs.Get(ctx, apiproject.ProjectID(req.ProjectId), path)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	if v == nil {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	resp := pb.GetResponse{
		Path:      v.Path.PB(),
		FetchedAt: timestamppb.New(v.FetchedAt),
	}

	if !req.OmitSource {
		resp.Source = v.Source
	}

	resp.Path.Version = v.FetchedVersion

	return &resp, nil
}
