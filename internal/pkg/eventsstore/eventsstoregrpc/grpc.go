package eventsstoregrpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbeventsvc "go.autokitteh.dev/idl/go/eventsvc"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
)

type Store struct{ Client pbeventsvc.EventsClient }

var _ eventsstore.Store = &Store{}

func (s *Store) Add(ctx context.Context, srcid apieventsrc.EventSourceID, assoc, originalID, typ string, data map[string]*apivalues.Value, memo map[string]string) (apievent.EventID, error) {
	resp, err := s.Client.IngestEvent(
		ctx,
		&pbeventsvc.IngestEventRequest{
			SrcId:            srcid.String(),
			Type:             typ,
			Memo:             memo,
			OriginalId:       originalID,
			AssociationToken: assoc,
			Data:             apivalues.StringValueMapToProto(data),
		},
	)
	if err != nil {
		return "", fmt.Errorf("ingest: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return "", fmt.Errorf("resp validate: %w", err)
	}

	return apievent.EventID(resp.Id), nil
}

func (s *Store) Get(ctx context.Context, id apievent.EventID) (*apievent.Event, error) {
	resp, err := s.Client.GetEvent(
		ctx,
		&pbeventsvc.GetEventRequest{
			Id: id.String(),
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, eventsstore.ErrNotFound
			}
		}

		return nil, fmt.Errorf("get: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	e, err := apievent.EventFromProto(resp.Event)
	if err != nil {
		return nil, fmt.Errorf("event: %w", err)
	}

	return e, nil
}

func (s *Store) GetStateForProject(ctx context.Context, id apievent.EventID, pid apiproject.ProjectID) ([]*apievent.ProjectEventStateRecord, error) {
	resp, err := s.Client.GetEventStateForProject(
		ctx,
		&pbeventsvc.GetEventStateForProjectRequest{
			Id:        id.String(),
			ProjectId: pid.String(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get state: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	rs := make([]*apievent.ProjectEventStateRecord, len(resp.Log))
	for i, l := range resp.Log {
		if rs[i], err = apievent.ProjectEventStateRecordFromProto(l); err != nil {
			return nil, fmt.Errorf("log %d: %w", i, err)
		}
	}

	return rs, nil
}

func (s *Store) UpdateStateForProject(ctx context.Context, id apievent.EventID, pid apiproject.ProjectID, state *apievent.ProjectEventState) error {
	resp, err := s.Client.UpdateEventStateForProject(
		ctx,
		&pbeventsvc.UpdateEventStateForProjectRequest{
			Id:        id.String(),
			ProjectId: pid.String(),
			State:     state.PB(),
		},
	)
	if err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return fmt.Errorf("resp validate: %w", err)
	}

	return nil
}

func (s *Store) List(ctx context.Context, pid *apiproject.ProjectID, ofs, l uint32) ([]*eventsstore.ListRecord, error) {
	if pid == nil {
		x := apiproject.ProjectID("")
		pid = &x
	}

	resp, err := s.Client.ListEvents(
		ctx,
		&pbeventsvc.ListEventsRequest{
			ProjectId: pid.String(),
			Ofs:       ofs,
			Len:       l,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("update state: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	rs := make([]*eventsstore.ListRecord, len(resp.Records))
	for i, pbr := range resp.Records {
		event, err := apievent.EventFromProto(pbr.Event)
		if err != nil {
			return nil, fmt.Errorf("#%d event: %w", i, err)
		}

		states := make([]*apievent.EventStateRecord, len(pbr.States))
		for j, pbs := range pbr.States {
			if states[j], err = apievent.EventStateRecordFromProto(pbs); err != nil {
				return nil, fmt.Errorf("#%d.%d state: %w", i, j, err)
			}
		}

		rs[i] = &eventsstore.ListRecord{
			Event:  event,
			States: states,
		}
	}

	return rs, nil
}

func (s *Store) UpdateState(ctx context.Context, id apievent.EventID, state *apievent.EventState) error {
	resp, err := s.Client.UpdateEventState(
		ctx,
		&pbeventsvc.UpdateEventStateRequest{
			Id:    id.String(),
			State: state.PB(),
		},
	)
	if err != nil {
		return fmt.Errorf("update state: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return fmt.Errorf("resp validate: %w", err)
	}

	return nil
}

func (s *Store) GetProjectWaitingEvents(ctx context.Context, pid apiproject.ProjectID) ([]apievent.EventID, error) {
	resp, err := s.Client.GetProjectWaitingEvents(ctx, &pbeventsvc.GetProjectWaitingEventsRequest{ProjectId: pid.String()})
	if err != nil {
		return nil, fmt.Errorf("get project waiting events: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	eids := make([]apievent.EventID, len(resp.EventIds))
	for i, id := range resp.EventIds {
		eids[i] = apievent.EventID(id)
	}

	return eids, nil
}

func (s *Store) GetState(ctx context.Context, id apievent.EventID) ([]*apievent.EventStateRecord, error) {
	resp, err := s.Client.GetEventState(
		ctx,
		&pbeventsvc.GetEventStateRequest{
			Id: id.String(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get state: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	rs := make([]*apievent.EventStateRecord, len(resp.Log))
	for i, l := range resp.Log {
		if rs[i], err = apievent.EventStateRecordFromProto(l); err != nil {
			return nil, fmt.Errorf("log %d: %w", i, err)
		}
	}

	return rs, nil
}

func (s *Store) Setup(context.Context) error { return fmt.Errorf("not supported") }

func (s *Store) Teardown(context.Context) error { return fmt.Errorf("not supported") }
