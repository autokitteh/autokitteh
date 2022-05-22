package fseventsrcsvc

import (
	"context"
	"errors"

	"github.com/fsnotify/fsnotify"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"

	pb "github.com/autokitteh/autokitteh/api/gen/stubs/go/fseventsrc"

	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/fseventsrc"
	"github.com/autokitteh/L"
)

type Svc struct {
	pb.UnimplementedFSEventSourceServer

	Src *fseventsrc.FSEventSource
	L   L.Nullable
}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pb.RegisterFSEventSourceServer(srv, s)

	if gw != nil {
		if err := pb.RegisterFSEventSourceHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Bind(ctx context.Context, req *pb.BindRequest) (*pb.BindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	var mask fsnotify.Op

	if req.Ops.Create {
		mask |= fsnotify.Create
	}

	if req.Ops.Write {
		mask |= fsnotify.Write
	}

	if req.Ops.Remove {
		mask |= fsnotify.Remove
	}

	if req.Ops.Rename {
		mask |= fsnotify.Rename
	}

	if req.Ops.Chmod {
		mask |= fsnotify.Chmod
	}

	if mask == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "at least a single op must be set")
	}

	if err := s.Src.Add(ctx, apiproject.ProjectID(req.ProjectId), req.Name, req.Path, mask); err != nil {
		return nil, status.Errorf(codes.Unknown, "add: %v", err)
	}

	return &pb.BindResponse{}, nil
}

func (s *Svc) Unbind(ctx context.Context, req *pb.UnbindRequest) (*pb.UnbindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Src.Remove(ctx, apiproject.ProjectID(req.ProjectId), req.Name, ""); err != nil {
		if errors.Is(err, fseventsrc.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "remove: %v", err)
	}

	return nil, nil
}
