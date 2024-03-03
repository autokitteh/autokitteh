package sdktypes

import (
	"go.jetpack.io/typeid"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type ExecutorID struct{ id[typeid.AnyPrefix] }

type concreteExecutorID interface {
	RunID | IntegrationID
	ID
}

func NewExecutorID[T concreteExecutorID](in T) ExecutorID {
	parsed := kittehs.Must1(ParseID[id[typeid.AnyPrefix]](in.String()))
	return ExecutorID{parsed}
}

func ParseExecutorID(s string) (ExecutorID, error) {
	parsed, err := ParseID[id[typeid.AnyPrefix]](s)
	if err != nil {
		return ExecutorID{}, err
	}

	switch parsed.Kind() {
	case runIDKind, integrationIDKind:
		return ExecutorID{parsed}, nil
	default:
		return ExecutorID{}, sdkerrors.NewInvalidArgumentError("invalid executor id")
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
