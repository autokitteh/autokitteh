package secretsstoregrpcsvc

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autokitteh/autokitteh/gen/proto/stubs/go/secretssvc"

	"github.com/autokitteh/autokitteh/internal/pkg/secretsstore"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Svc struct {
	pb.UnimplementedSecretsServer

	Store *secretsstore.Store

	L L.Nullable
}

var _ pb.SecretsServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pb.RegisterSecretsServer(srv, s)

	if gw != nil {
		if err := pb.RegisterSecretsHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Store.Set(ctx, apiproject.ProjectID(req.ProjectId), req.Name, req.Value); err != nil {
		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	return &pb.SetResponse{}, nil
}

func (s *Svc) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	v, err := s.Store.Get(ctx, apiproject.ProjectID(req.ProjectId), req.Name)
	if err != nil {
		if errors.Is(err, secretsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	return &pb.GetResponse{Value: v}, nil
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
