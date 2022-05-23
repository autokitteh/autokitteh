package eventsstoregorm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
)

type Store struct {
	DB *gorm.DB
}

var _ eventsstore.Store = &Store{}

func (s *Store) Add(ctx context.Context, srcid apieventsrc.EventSourceID, assoc, originalID, typ string, data map[string]*apivalues.Value, memo map[string]string) (apievent.EventID, error) {
	id := apievent.NewEventID()

	e, err := apievent.NewEvent(id, srcid, assoc, originalID, typ, data, memo, time.Now())
	if err != nil {
		return "", fmt.Errorf("new Event: %w", err)
	}

	r, err := encodeEvent(e)
	if err != nil {
		return "", fmt.Errorf("Event: %w", err)
	}

	if err := s.DB.WithContext(ctx).Create(&r).Error; err != nil {
		return "", fmt.Errorf("create: %w", err)
	}

	return id, nil
}

func (s *Store) Get(ctx context.Context, id apievent.EventID) (*apievent.Event, error) {
	var r Event

	if err := s.DB.WithContext(ctx).First(&r, "id = ?", id.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, eventsstore.ErrNotFound
		}

		return nil, err
	}

	e, err := decodeEvent(&r)
	if err != nil {
		return nil, fmt.Errorf("Event: %w", err)
	}

	return e, nil
}

func (s *Store) UpdateStateForProject(ctx context.Context, eid apievent.EventID, pid apiproject.ProjectID, es *apievent.ProjectEventState) error {
	esr, err := apievent.NewProjectEventStateRecord(es, time.Now())
	if err != nil {
		return fmt.Errorf("state record: %w", err)
	}

	r, err := encodeProjectEventState(eid, pid, esr)
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}

	if err := s.DB.WithContext(ctx).Create(&r).Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (s *Store) GetStateForProject(ctx context.Context, eid apievent.EventID, pid apiproject.ProjectID) ([]*apievent.ProjectEventStateRecord, error) {
	var esrs []projectEventState

	if err := s.DB.
		WithContext(ctx).
		Order("t desc").
		Where("event_id = ?", eid.String()).
		Where("project_id = ?", pid.String()).
		Find(&esrs).Error; err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	es := make([]*apievent.ProjectEventStateRecord, len(esrs))
	for i, esr := range esrs {
		var err error
		if es[i], err = decodeProjectEventState(&esr); err != nil {
			return nil, fmt.Errorf("record %d: %w", i, err)
		}
	}

	return es, nil
}

func (s *Store) List(ctx context.Context, pid *apiproject.ProjectID, ofs, l uint32) ([]*eventsstore.ListRecord, error) {
	q := s.DB.WithContext(ctx).Model(&Event{})

	// TODO: allow Event and EventState to be internal.
	type rec struct {
		Event
		EventState
	}

	var rs []rec

	if pid != nil {
		q = q.
			Joins("inner join project_event_states on events.id == project_event_states.event_id").
			Where("project_id = ?", pid)
	}

	if l == 0 {
		l = math.MaxUint32
	}

	q = q.
		Joins("inner join event_states on events.id == event_states.event_id").
		// limit and offset number of total events.
		Joins("inner join (select events.id from events order by created_at desc limit ? offset ?) ids on ids.id == events.id", l, ofs)

	if err := q.Order("created_at desc").Select("*").Find(&rs).Error; err != nil {
		return nil, err
	}

	eids := make(map[string]int, len(rs)) // for events deduping.

	lrs := make([]*eventsstore.ListRecord, 0, len(rs))
	for i, r := range rs {
		var lr *eventsstore.ListRecord

		j, ok := eids[r.ID]
		if ok {
			lr = lrs[j]
		} else {
			lr = &eventsstore.ListRecord{}

			var err error
			if lr.Event, err = decodeEvent(&r.Event); err != nil {
				return nil, fmt.Errorf("#%d Event: %w", i, err)
			}

			lrs = append(lrs, lr)
			eids[r.ID] = len(lrs) - 1
		}

		state, err := decodeEventState(&r.EventState)
		if err != nil {
			return nil, fmt.Errorf("#%d state: %w", i, err)
		}

		// TODO: this is a hack since we get redundant project state records for each
		// event. Need to make the SQL query handle this gracefully somehow.
		if len(lr.States) == 0 || state.T().After(lr.States[len(lr.States)-1].T()) {
			lr.States = append([]*apievent.EventStateRecord{state}, lr.States...)
		}
	}

	return lrs, nil
}

func (s *Store) UpdateState(ctx context.Context, id apievent.EventID, es *apievent.EventState) error {
	esr, err := apievent.NewEventStateRecord(es, time.Now())
	if err != nil {
		return fmt.Errorf("state record: %w", err)
	}

	r, err := encodeEventState(id, esr)
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}

	if err := s.DB.WithContext(ctx).Create(&r).Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (s *Store) GetState(ctx context.Context, id apievent.EventID) ([]*apievent.EventStateRecord, error) {
	var esrs []EventState

	if err := s.DB.
		WithContext(ctx).
		Order("t desc").
		Where("event_id = ?", id.String()).
		Find(&esrs).Error; err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	es := make([]*apievent.EventStateRecord, len(esrs))
	for i, esr := range esrs {
		var err error
		if es[i], err = decodeEventState(&esr); err != nil {
			return nil, fmt.Errorf("record %d: %w", i, err)
		}
	}

	return es, nil
}

func (s *Store) GetProjectWaitingEvents(ctx context.Context, pid apiproject.ProjectID) ([]apievent.EventID, error) {
	var ids []string

	if err := s.DB.
		WithContext(ctx).
		Model(projectEventState{}).
		Where("project_id = ?", pid.String()).
		Where("state_name = ?", "waiting").
		Pluck("event_id", &ids).
		Error; err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	eids := make([]apievent.EventID, len(ids))
	for i, id := range ids {
		eids[i] = apievent.EventID(id)
	}

	return eids, nil
}

func (s *Store) Setup(ctx context.Context) error {
	if err := s.DB.WithContext(ctx).AutoMigrate(&Event{}, &projectEventState{}, &EventState{}); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	return nil
}

func (s *Store) Teardown(ctx context.Context) error {
	if err := s.DB.WithContext(ctx).Migrator().DropTable(&Event{}, &projectEventState{}, &EventState{}); err != nil {
		return fmt.Errorf("drop: %w", err)
	}

	return nil
}
