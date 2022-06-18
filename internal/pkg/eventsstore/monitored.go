package eventsstore

import (
	"context"
	"time"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
)

type MonitoredStore struct {
	Store Store

	EventStateUpdate        func(apievent.EventID, *apievent.EventStateRecord /* nil = just added */)
	ProjectEventStateUpdate func(apievent.EventID, apiproject.ProjectID, *apievent.ProjectEventStateRecord)
}

func (m *MonitoredStore) Add(
	ctx context.Context,
	id apievent.EventID,
	srcid apieventsrc.EventSourceID,
	assoc string,
	originalID string,
	typ string,
	data map[string]*apivalues.Value,
	memo map[string]string,
) (apievent.EventID, error) {
	id, err := m.Store.Add(ctx, id, srcid, assoc, originalID, typ, data, memo)

	if err == nil {
		go m.EventStateUpdate(id, nil)
	}

	return id, err
}

func (m *MonitoredStore) Get(ctx context.Context, id apievent.EventID) (*apievent.Event, error) {
	return m.Store.Get(ctx, id)
}

func (m *MonitoredStore) UpdateState(ctx context.Context, id apievent.EventID, s *apievent.EventState) error {
	if err := m.Store.UpdateState(ctx, id, s); err != nil {
		return err
	}

	go m.EventStateUpdate(id, apievent.MustNewEventStateRecord(s, time.Now()))

	return nil
}

func (m *MonitoredStore) GetState(ctx context.Context, id apievent.EventID) ([]*apievent.EventStateRecord, error) {
	return m.Store.GetState(ctx, id)
}

func (m *MonitoredStore) GetStateForProject(ctx context.Context, eid apievent.EventID, pid apiproject.ProjectID) ([]*apievent.ProjectEventStateRecord, error) {
	return m.Store.GetStateForProject(ctx, eid, pid)
}

func (m *MonitoredStore) UpdateStateForProject(ctx context.Context, eid apievent.EventID, pid apiproject.ProjectID, s *apievent.ProjectEventState) error {
	if err := m.Store.UpdateStateForProject(ctx, eid, pid, s); err != nil {
		return err
	}

	go m.ProjectEventStateUpdate(eid, pid, apievent.MustNewProjectEventStateRecord(s, time.Now()))

	return nil
}

func (m *MonitoredStore) GetProjectWaitingEvents(ctx context.Context, pid apiproject.ProjectID) ([]apievent.EventID, error) {
	return m.Store.GetProjectWaitingEvents(ctx, pid)
}

func (m *MonitoredStore) List(ctx context.Context, pid *apiproject.ProjectID, ofs, l uint32) ([]*ListRecord, error) {
	return m.Store.List(ctx, pid, ofs, l)
}

func (m *MonitoredStore) Setup(ctx context.Context) error {
	return m.Store.Setup(ctx)
}

func (m *MonitoredStore) Teardown(ctx context.Context) error {
	return m.Store.Teardown(ctx)
}
