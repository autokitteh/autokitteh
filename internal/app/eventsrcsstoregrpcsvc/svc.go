package eventsrcsstoregrpcsvc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbeventsrc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/eventsrc"
	pbeventsrcsvc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/eventsrcsvc"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Svc struct {
	pbeventsrcsvc.UnimplementedEventSourcesServer

	Store eventsrcsstore.Store

	L L.Nullable
}

var _ pbeventsrcsvc.EventSourcesServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server) {
	pbeventsrcsvc.RegisterEventSourcesServer(srv, s)
}

func (s *Svc) AddEventSource(ctx context.Context, req *pbeventsrcsvc.AddEventSourceRequest) (*pbeventsrcsvc.AddEventSourceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apieventsrc.EventSourceSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	if err := s.Store.Add(ctx, apieventsrc.EventSourceID(req.Id), d); err != nil {
		if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return nil, status.Errorf(codes.AlreadyExists, "account of %s", req.Id)
		} else if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "account of %s", req.Id)
		}

		return nil, status.Errorf(codes.Unknown, "add: %v", err)
	}

	return &pbeventsrcsvc.AddEventSourceResponse{}, nil
}

func (s *Svc) UpdateEventSource(ctx context.Context, req *pbeventsrcsvc.UpdateEventSourceRequest) (*pbeventsrcsvc.UpdateEventSourceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apieventsrc.EventSourceSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	err = s.Store.Update(ctx, apieventsrc.EventSourceID(req.Id), d)
	if err != nil {
		if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "update: %v", err)
	}

	return &pbeventsrcsvc.UpdateEventSourceResponse{}, nil
}

func (s *Svc) GetEventSource(ctx context.Context, req *pbeventsrcsvc.GetEventSourceRequest) (*pbeventsrcsvc.GetEventSourceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	src, err := s.Store.Get(ctx, apieventsrc.EventSourceID(req.Id))
	if err != nil {
		if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "update: %v", err)
	}

	return &pbeventsrcsvc.GetEventSourceResponse{Src: src.PB()}, nil
}

func (s *Svc) AddEventSourceProjectBinding(ctx context.Context, req *pbeventsrcsvc.AddEventSourceProjectBindingRequest) (*pbeventsrcsvc.AddEventSourceProjectBindingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apieventsrc.EventSourceProjectBindingSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	err = s.Store.AddProjectBinding(
		ctx,
		apieventsrc.EventSourceID(req.SrcId),
		apiproject.ProjectID(req.ProjectId),
		req.Name,
		req.AssociationToken,
		req.SourceConfig,
		req.Approved,
		d,
	)
	if err != nil {
		if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return nil, status.Errorf(codes.AlreadyExists, "already exists")
		} else if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "add: %v", err)
	}

	return &pbeventsrcsvc.AddEventSourceProjectBindingResponse{}, nil
}

func (s *Svc) UpdateEventSourceProjectBinding(ctx context.Context, req *pbeventsrcsvc.UpdateEventSourceProjectBindingRequest) (*pbeventsrcsvc.UpdateEventSourceProjectBindingResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apieventsrc.EventSourceProjectBindingSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	err = s.Store.UpdateProjectBinding(
		ctx,
		apieventsrc.EventSourceID(req.SrcId),
		apiproject.ProjectID(req.ProjectId),
		req.Name,
		req.Approved,
		d,
	)
	if err != nil {
		if errors.Is(err, eventsrcsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "update: %v", err)
	}

	return &pbeventsrcsvc.UpdateEventSourceProjectBindingResponse{}, nil
}

func (s *Svc) GetEventSourceProjectBindings(ctx context.Context, req *pbeventsrcsvc.GetEventSourceProjectBindingsRequest) (*pbeventsrcsvc.GetEventSourceProjectBindingsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	var (
		srcid *apieventsrc.EventSourceID
		_pid  = apiproject.ProjectID(req.ProjectId)
		pid   *apiproject.ProjectID
	)

	if _pid.String() != "" {
		pid = &_pid
	}

	if req.Id != "" {
		srcid_ := apieventsrc.EventSourceID(req.Id)
		srcid = &srcid_
	}

	bs, err := s.Store.GetProjectBindings(ctx, srcid, pid, req.Name, req.AssociationToken, !req.IncludeUnapproved)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get bindings: %v", err)
	}

	pbbs := make([]*pbeventsrc.EventSourceProjectBinding, len(bs))
	for i, b := range bs {
		pbbs[i] = b.PB()
	}

	return &pbeventsrcsvc.GetEventSourceProjectBindingsResponse{Bindings: pbbs}, nil
}

func (s *Svc) ListEventSources(ctx context.Context, req *pbeventsrcsvc.ListEventSourcesRequest) (*pbeventsrcsvc.ListEventSourcesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	var (
		_aname = apiaccount.AccountName(req.AccountName)
		aname  *apiaccount.AccountName
	)

	if _aname.String() != "" {
		aname = &_aname
	}

	ids, err := s.Store.List(ctx, aname)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "list: %v", err)
	}

	pbids := make([]string, len(ids))
	for i, id := range ids {
		pbids[i] = id.String()
	}

	return &pbeventsrcsvc.ListEventSourcesResponse{Ids: pbids}, nil
}
