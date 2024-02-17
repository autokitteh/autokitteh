package sdkexecutor

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Executor interface {
	ExecutorID() sdktypes.ExecutorID

	Values() map[string]sdktypes.Value

	Caller
}

type executor struct {
	Caller

	xid sdktypes.ExecutorID
	vs  map[string]sdktypes.Value
}

func (m *executor) ExecutorID() sdktypes.ExecutorID   { return m.xid }
func (m *executor) Values() map[string]sdktypes.Value { return m.vs }

func NewExecutor(caller Caller, xid sdktypes.ExecutorID, vs map[string]sdktypes.Value) Executor {
	return &executor{Caller: caller, xid: xid, vs: vs}
}
