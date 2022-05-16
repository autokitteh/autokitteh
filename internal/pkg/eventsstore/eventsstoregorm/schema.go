package eventsstoregorm

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apievent"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
)

type Event struct {
	ID          string    `gorm:"primaryKey;index:idx_ord"`
	SrcID       string    `gorm:"not null;index:idx_ord"`
	CreatedAt   time.Time `gorm:"not null;index:idx_ord"`
	Type        string    `gorm:"not null"`
	OriginalID  string    `gorm:"index"`
	Association string
	Data        datatypes.JSON
	Memo        datatypes.JSON
}

func decodeEvent(e *Event) (*apievent.Event, error) {
	var (
		data map[string]*apivalues.Value
		memo map[string]string
	)

	if err := json.Unmarshal(e.Data, &data); err != nil {
		return nil, fmt.Errorf("data: %w", err)
	}

	if err := json.Unmarshal(e.Memo, &memo); err != nil {
		return nil, fmt.Errorf("memo: %w", err)
	}

	return apievent.NewEvent(
		apievent.EventID(e.ID),
		apieventsrc.EventSourceID(e.SrcID),
		e.Association,
		e.OriginalID,
		e.Type,
		data,
		memo,
		e.CreatedAt,
	)
}

func encodeEvent(e *apievent.Event) (*Event, error) {
	data, err := json.Marshal(e.Data())
	if err != nil {
		return nil, fmt.Errorf("data: %w", err)
	}

	memo, err := json.Marshal(e.Memo())
	if err != nil {
		return nil, fmt.Errorf("memo: %w", err)
	}

	return &Event{
		ID:          e.ID().String(),
		SrcID:       e.EventSourceID().String(),
		CreatedAt:   e.T(),
		Type:        e.Type(),
		Data:        datatypes.JSON(data),
		Memo:        datatypes.JSON(memo),
		Association: e.AssociationToken(),
		OriginalID:  e.OriginalID(),
	}, nil
}

//--

type projectEventState struct {
	EventID   string         `gorm:"index:idx_pord;not null"`
	ProjectID string         `gorm:"index:idx_pord;index:idx_states;not null"`
	T         time.Time      `gorm:"index:idx_pord;not null"`
	StateName string         `gorm:"index:idx_states;not null"`
	State     datatypes.JSON `gorm:"not null"`
}

func decodeProjectEventState(p *projectEventState) (*apievent.ProjectEventStateRecord, error) {
	var state apievent.ProjectEventState
	if err := json.Unmarshal(p.State, &state); err != nil {
		return nil, fmt.Errorf("state: %w", err)
	}

	return apievent.NewProjectEventStateRecord(
		&state,
		p.T,
	)
}

func encodeProjectEventState(
	eid apievent.EventID,
	pid apiproject.ProjectID,
	r *apievent.ProjectEventStateRecord,
) (*projectEventState, error) {
	state, err := json.Marshal(r.State())
	if err != nil {
		return nil, fmt.Errorf("project state: %w", err)
	}

	return &projectEventState{
		EventID:   eid.String(),
		ProjectID: pid.String(),
		T:         r.T(),
		StateName: r.State().Name(),
		State:     state,
	}, nil
}

//--

type EventState struct {
	EventID string         `gorm:"index:idx_sord;not null"`
	T       time.Time      `gorm:"index:idx_sord;not null"`
	State   datatypes.JSON `gorm:"not null"`
}

func decodeEventState(p *EventState) (*apievent.EventStateRecord, error) {
	var state apievent.EventState
	if err := json.Unmarshal(p.State, &state); err != nil {
		return nil, fmt.Errorf("state: %w", err)
	}

	return apievent.NewEventStateRecord(
		&state,
		p.T,
	)
}

func encodeEventState(
	eid apievent.EventID,
	r *apievent.EventStateRecord,
) (*EventState, error) {
	state, err := json.Marshal(r.State())
	if err != nil {
		return nil, fmt.Errorf("state: %w", err)
	}

	return &EventState{
		EventID: eid.String(),
		T:       r.T(),
		State:   state,
	}, nil
}
