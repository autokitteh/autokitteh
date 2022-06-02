package eventsstore

import (
	"context"
	"errors"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
)

var ErrNotFound = errors.New("not found")

type ListRecord struct {
	Event  *apievent.Event
	States []*apievent.EventStateRecord
}

type Store interface {
	Add(
		_ context.Context,
		_ apievent.EventID,
		_ apieventsrc.EventSourceID,
		assoc string,
		originalID string,
		typ string,
		_ map[string]*apivalues.Value,
		_ map[string]string,
	) (apievent.EventID, error)
	Get(context.Context, apievent.EventID) (*apievent.Event, error)
	UpdateState(context.Context, apievent.EventID, *apievent.EventState) error
	GetState(context.Context, apievent.EventID) ([]*apievent.EventStateRecord, error)

	// TODO: optional project id, will list for all projects.
	GetStateForProject(context.Context, apievent.EventID, apiproject.ProjectID) ([]*apievent.ProjectEventStateRecord, error)
	UpdateStateForProject(context.Context, apievent.EventID, apiproject.ProjectID, *apievent.ProjectEventState) error

	GetProjectWaitingEvents(context.Context, apiproject.ProjectID) ([]apievent.EventID, error)

	List(_ context.Context, _ *apiproject.ProjectID, ofs, l uint32) ([]*ListRecord, error)

	Setup(context.Context) error
	Teardown(context.Context) error
}
