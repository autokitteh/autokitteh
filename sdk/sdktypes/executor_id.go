package sdktypes

import (
	"go.jetify.com/typeid"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type ExecutorID struct{ id[typeid.AnyPrefix] }

var InvalidExecutorID ExecutorID

type concreteExecutorID interface {
	RunID | IntegrationID
	ID
}

func NewExecutorID[T concreteExecutorID](in T) ExecutorID {
	parsed := typeid.Must(ParseID[id[typeid.AnyPrefix]](in.String()))
	return ExecutorID{parsed}
}

func ParseExecutorID(s string) (ExecutorID, error) {
	if s == "" {
		return InvalidExecutorID, nil
	}

	parsed, err := ParseID[id[typeid.AnyPrefix]](s)
	if err != nil {
		return InvalidExecutorID, err
	}

	switch parsed.Kind() {
	case runIDKind, integrationIDKind:
		return ExecutorID{parsed}, nil
	default:
		return InvalidExecutorID, sdkerrors.NewInvalidArgumentError("invalid executor id")
	}
}

func (e ExecutorID) ToRunID() RunID {
	id, _ := ParseRunID(e.String())
	return id
}

func (e ExecutorID) ToIntegrationID() IntegrationID {
	id, _ := ParseIntegrationID(e.String())
	return id
}

func (e ExecutorID) IsRunID() bool         { return e.Kind() == runIDKind }
func (e ExecutorID) IsIntegrationID() bool { return e.Kind() == integrationIDKind }
