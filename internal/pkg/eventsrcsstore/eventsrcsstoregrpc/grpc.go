package eventsrcsstoregrpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbeventsrcsvc "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/eventsrcsvc"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsrcsstore"
)

type Store struct {
	Client pbeventsrcsvc.EventSourcesClient
}

var _ eventsrcsstore.Store = &Store{}

func (s *Store) Add(ctx context.Context, sid apieventsrc.EventSourceID, data *apieventsrc.EventSourceSettings) error {
	resp, err := s.Client.AddEventSource(ctx, &pbeventsrcsvc.AddEventSourceRequest{Id: sid.String(), Settings: data.PB()})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				return eventsrcsstore.ErrAlreadyExists
			} else if e.Code() == codes.NotFound {
				return eventsrcsstore.ErrNotFound
			}
		}

		return fmt.Errorf("add: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return fmt.Errorf("resp: %w", err)
	}

	return nil
}

func (s *Store) Get(ctx context.Context, id apieventsrc.EventSourceID) (*apieventsrc.EventSource, error) {
	resp, err := s.Client.GetEventSource(ctx, &pbeventsrcsvc.GetEventSourceRequest{Id: id.String()})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, eventsrcsstore.ErrNotFound
			}
		}

		return nil, fmt.Errorf("get: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp: %w", err)
	}

	src, err := apieventsrc.EventSourceFromProto(resp.Src)
	if err != nil {
		return nil, fmt.Errorf("src: %w", err)
	}

	return src, err
}

func (s *Store) Update(ctx context.Context, sid apieventsrc.EventSourceID, data *apieventsrc.EventSourceSettings) error {
	_, err := s.Client.UpdateEventSource(
		ctx,
		&pbeventsrcsvc.UpdateEventSourceRequest{
			Id:       sid.String(),
			Settings: data.PB(),
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return eventsrcsstore.ErrNotFound
			}
		}

		return fmt.Errorf("update: %w", err)
	}

	return nil
}

func (s *Store) List(ctx context.Context, aname *apiaccount.AccountName) ([]apieventsrc.EventSourceID, error) {
	resp, err := s.Client.ListEventSources(
		ctx,
		&pbeventsrcsvc.ListEventSourcesRequest{
			AccountName: aname.MaybeString(),
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, eventsrcsstore.ErrNotFound
			}
		}

		return nil, fmt.Errorf("list: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp: %w", err)
	}

	ids := make([]apieventsrc.EventSourceID, len(resp.Ids))
	for i, id := range resp.Ids {
		ids[i] = apieventsrc.EventSourceID(id)
	}

	return ids, nil
}

func (s *Store) AddProjectBinding(ctx context.Context, srcid apieventsrc.EventSourceID, pid apiproject.ProjectID, name string, assoc, cfg string, approved bool, data *apieventsrc.EventSourceProjectBindingSettings) error {
	_, err := s.Client.AddEventSourceProjectBinding(
		ctx,
		&pbeventsrcsvc.AddEventSourceProjectBindingRequest{
			SrcId:            srcid.String(),
			Name:             name,
			AssociationToken: assoc,
			SourceConfig:     cfg,
			ProjectId:        pid.String(),
			Approved:         approved,
			Settings:         data.PB(),
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				return eventsrcsstore.ErrAlreadyExists
			} else if e.Code() == codes.NotFound {
				return eventsrcsstore.ErrNotFound
			}
		}

		return fmt.Errorf("add binding: %w", err)
	}

	return nil
}

func (s *Store) UpdateProjectBinding(ctx context.Context, srcid apieventsrc.EventSourceID, pid apiproject.ProjectID, name string, approved bool, data *apieventsrc.EventSourceProjectBindingSettings) error {
	_, err := s.Client.UpdateEventSourceProjectBinding(
		ctx,
		&pbeventsrcsvc.UpdateEventSourceProjectBindingRequest{
			SrcId:     srcid.String(),
			ProjectId: pid.String(),
			Name:      name,
			Approved:  approved,
			Settings:  data.PB(),
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return eventsrcsstore.ErrNotFound
			}
		}

		return fmt.Errorf("add binding: %w", err)
	}

	return nil
}

func (s *Store) GetProjectBindings(ctx context.Context, srcid *apieventsrc.EventSourceID, pid *apiproject.ProjectID, name, assoc string, approvedOnly bool) ([]*apieventsrc.EventSourceProjectBinding, error) {
	resp, err := s.Client.GetEventSourceProjectBindings(
		ctx,
		&pbeventsrcsvc.GetEventSourceProjectBindingsRequest{
			Id:                srcid.String(),
			ProjectId:         pid.MaybeString(),
			Name:              name,
			IncludeUnapproved: !approvedOnly,
			AssociationToken:  assoc,
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, eventsrcsstore.ErrNotFound
			}
		}

		return nil, fmt.Errorf("get binding: %w", err)
	}

	bs := make([]*apieventsrc.EventSourceProjectBinding, len(resp.Bindings))
	for i, b := range resp.Bindings {
		if bs[i], err = apieventsrc.EventSourceProjectBindingFromProto(b); err != nil {
			return nil, fmt.Errorf("binding %d: %w", i, err)
		}
	}

	return bs, nil
}

func (s *Store) Setup(ctx context.Context) error { return fmt.Errorf("not supported") }

func (s *Store) Teardown(ctx context.Context) error { return fmt.Errorf("not supported") }
