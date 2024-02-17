package sdktypes

import (
	"encoding/json"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type ExecutorID = *executorID

type executorID struct{ id ID }

var _ ID = (ExecutorID)(nil)

func (id *executorID) isID()         {}
func (id *executorID) Kind() string  { return id.id.Kind() }
func (id *executorID) Value() string { return id.id.Value() }

func (id *executorID) MarshalJSON() ([]byte, error) {
	if id == nil {
		return []byte("null"), nil
	}

	return id.id.MarshalJSON()
}

func (id *executorID) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}

	id1, err := StrictParseExecutorID(text)
	if err != nil {
		return err
	}

	*id = *id1
	return nil
}

func (id *executorID) String() string {
	if id == nil {
		return ""
	}

	return id.id.String()
}

func (id *executorID) ToRunID() RunID {
	if id == nil {
		return nil
	}

	if rid, ok := id.id.(RunID); ok {
		return rid
	}

	return nil
}

func (id *executorID) ToIntegrationID() IntegrationID {
	if id == nil {
		return nil
	}

	if rid, ok := id.id.(IntegrationID); ok {
		return rid
	}

	return nil
}

func toExecutorID(id ID) ExecutorID {
	if id == nil {
		return nil
	}

	return &executorID{id: id}
}

func (id *executorID) ID() ID { return id.id }

type executorIDConstraint interface {
	RunID | IntegrationID
	ID
}

func NewExecutorID[T executorIDConstraint](id T) ExecutorID {
	if id == nil {
		return nil
	}
	return &executorID{id: id}
}

func StrictParseExecutorID(raw string) (ExecutorID, error) {
	k, _, _ := SplitRawID(raw)

	if k == RunIDKind {
		rid, err := ParseRunID(raw)
		return toExecutorID(rid), err
	}

	if k == IntegrationIDKind {
		rid, err := ParseIntegrationID(raw)
		return toExecutorID(rid), err
	}

	if k == ConnectionIDKind {
		rid, err := ParseConnectionID(raw)
		return toExecutorID(rid), err
	}

	return nil, sdkerrors.ErrInvalidArgument
}

func ParseExecutorID(raw string) (ExecutorID, error) {
	if raw == "" {
		return nil, nil
	}

	return StrictParseExecutorID(raw)
}

func MustParseExecutorID(raw string) ExecutorID {
	return kittehs.Must1(ParseExecutorID(raw))
}
