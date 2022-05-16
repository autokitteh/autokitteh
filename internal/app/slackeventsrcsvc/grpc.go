package slackeventsrcsvc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autokitteh/autokitteh/gen/proto/stubs/go/slackeventsrc"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
)

func (s *Svc) Bind(ctx context.Context, req *pb.BindRequest) (*pb.BindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Add(ctx, apiproject.ProjectID(req.ProjectId), req.Name, req.TeamId); err != nil {
		return nil, status.Errorf(codes.Unknown, "add: %v", err)
	}

	return &pb.BindResponse{}, nil
}

func (s *Svc) Unbind(ctx context.Context, req *pb.UnbindRequest) (*pb.UnbindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Remove(ctx, apiproject.ProjectID(req.ProjectId), req.Name); err != nil {
		return nil, status.Errorf(codes.Unknown, "remove: %v", err)
	}

	return &pb.UnbindResponse{}, nil
}
