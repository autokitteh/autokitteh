package statestoregrpcsvc

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/autokitteh/autokitteh/api/gen/stubs/go/statesvc"

	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
	L "github.com/autokitteh/L"
)

type Svc struct {
	pb.UnimplementedStateServer

	Store statestore.Store

	L L.Nullable
}

var _ pb.StateServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pb.RegisterStateServer(srv, s)

	if gw != nil {
		if err := pb.RegisterStateHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	v, err := apivalues.ValueFromProto(req.Value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "value: %v", err)
	}

	if err := s.Store.Set(ctx, apiproject.ProjectID(req.ProjectId), req.Name, v); err != nil {
		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	return &pb.SetResponse{}, nil
}

func (s *Svc) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	v, m, err := s.Store.Get(ctx, apiproject.ProjectID(req.ProjectId), req.Name)
	if err != nil {
		if errors.Is(err, statestore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	return &pb.GetResponse{
		Value: v.PB(),
		Metadata: &pb.ValueMetadata{
			UpdatedAt: timestamppb.New(m.UpdatedAt),
		},
	}, nil
}

func (s *Svc) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	ks, err := s.Store.List(ctx, apiproject.ProjectID(req.ProjectId))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	return &pb.ListResponse{Names: ks}, nil
}
