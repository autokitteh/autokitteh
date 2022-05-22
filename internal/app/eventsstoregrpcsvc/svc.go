package eventsstoregrpcsvc

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbevent "github.com/autokitteh/autokitteh/api/gen/stubs/go/event"
	pbeventsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/eventsvc"

	"github.com/autokitteh/autokitteh/sdk/api/apievent"
	"github.com/autokitteh/autokitteh/sdk/api/apieventsrc"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	L "github.com/autokitteh/L"
)

type Svc struct {
	pbeventsvc.UnimplementedEventsServer

	Events *events.Events
	Store  eventsstore.Store

	L L.Nullable
}

var _ pbeventsvc.EventsServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pbeventsvc.RegisterEventsServer(srv, s)

	if gw != nil {
		if err := pbeventsvc.RegisterEventsHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) IngestEvent(ctx context.Context, req *pbeventsvc.IngestEventRequest) (*pbeventsvc.IngestEventResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	data, err := apivalues.StringValueMapFromProto(req.Data)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	srcid := apieventsrc.EventSourceID(req.SrcId)

	var id apievent.EventID

	if s.Events != nil {
		id, err = s.Events.IngestEvent(ctx, srcid, req.AssociationToken, req.OriginalId, req.Type, data, req.Memo)
	} else {
		id, err = s.Store.Add(ctx, srcid, req.AssociationToken, req.OriginalId, req.Type, data, req.Memo)
	}

	if err != nil {
		return nil, status.Errorf(codes.Unknown, "add: %v", err)
	}

	return &pbeventsvc.IngestEventResponse{Id: id.String()}, nil
}

func (s *Svc) GetEvent(ctx context.Context, req *pbeventsvc.GetEventRequest) (*pbeventsvc.GetEventResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	ev, err := s.Store.Get(ctx, apievent.EventID(req.Id))
	if err != nil {
		if errors.Is(err, eventsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "get %s", req.Id)
		}

		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	return &pbeventsvc.GetEventResponse{Event: ev.PB()}, nil
}

func (s *Svc) ListEvents(ctx context.Context, req *pbeventsvc.ListEventsRequest) (*pbeventsvc.ListEventsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	var pid *apiproject.ProjectID
	if req.ProjectId != "" {
		rpid := apiproject.ProjectID(req.ProjectId)
		pid = &rpid
	}

	rs, err := s.Store.List(ctx, pid, req.Ofs, req.Len)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "list: %v", err)
	}

	pbrs := make([]*pbeventsvc.ListEventRecord, len(rs))
	for i, r := range rs {
		pbrs[i] = &pbeventsvc.ListEventRecord{
			Event:  r.Event.PB(),
			States: make([]*pbevent.EventStateRecord, len(r.States)),
		}

		for j, s := range r.States {
			pbrs[i].States[j] = s.PB()
		}
	}

	return &pbeventsvc.ListEventsResponse{Records: pbrs}, nil
}

func (s *Svc) GetEventStateForProject(ctx context.Context, req *pbeventsvc.GetEventStateForProjectRequest) (*pbeventsvc.GetEventStateForProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	states, err := s.Store.GetStateForProject(ctx, apievent.EventID(req.Id), apiproject.ProjectID(req.ProjectId))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get state: %v", err)
	}

	log := make([]*pbevent.ProjectEventStateRecord, len(states))
	for i, s := range states {
		log[i] = s.PB()
	}

	return &pbeventsvc.GetEventStateForProjectResponse{Log: log}, nil
}

func (s *Svc) UpdateEventStateForProject(ctx context.Context, req *pbeventsvc.UpdateEventStateForProjectRequest) (*pbeventsvc.UpdateEventStateForProjectResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	state, err := apievent.ProjectEventStateFromProto(req.State)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "state: %v", err)
	}

	if err := s.Store.UpdateStateForProject(
		ctx,
		apievent.EventID(req.Id),
		apiproject.ProjectID(req.ProjectId),
		state,
	); err != nil {
		return nil, status.Errorf(codes.Unknown, "update state: %v", err)
	}

	return &pbeventsvc.UpdateEventStateForProjectResponse{}, nil
}

func (s *Svc) GetEventState(ctx context.Context, req *pbeventsvc.GetEventStateRequest) (*pbeventsvc.GetEventStateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	states, err := s.Store.GetState(ctx, apievent.EventID(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get state: %v", err)
	}

	log := make([]*pbevent.EventStateRecord, len(states))
	for i, s := range states {
		log[i] = s.PB()
	}

	return &pbeventsvc.GetEventStateResponse{Log: log}, nil
}

func (s *Svc) UpdateEventState(ctx context.Context, req *pbeventsvc.UpdateEventStateRequest) (*pbeventsvc.UpdateEventStateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	state, err := apievent.EventStateFromProto(req.State)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "state: %v", err)
	}

	if err := s.Store.UpdateState(
		ctx,
		apievent.EventID(req.Id),
		state,
	); err != nil {
		return nil, status.Errorf(codes.Unknown, "update state: %v", err)
	}

	return &pbeventsvc.UpdateEventStateResponse{}, nil
}

func (s *Svc) GetProjectWaitingEvents(ctx context.Context, req *pbeventsvc.GetProjectWaitingEventsRequest) (*pbeventsvc.GetProjectWaitingEventsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	eids, err := s.Store.GetProjectWaitingEvents(ctx, apiproject.ProjectID(req.ProjectId))
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get project waiting events: %v", err)
	}

	ids := make([]string, len(eids))
	for i, id := range eids {
		ids[i] = string(id)
	}

	return &pbeventsvc.GetProjectWaitingEventsResponse{EventIds: ids}, nil
}
