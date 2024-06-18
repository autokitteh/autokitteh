package sdkexecutor

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Executor interface {
	ExecutorIDs() []sdktypes.ExecutorID

	Values() map[string]sdktypes.Value

	Caller
}

type executor struct {
	Caller

	xids []sdktypes.ExecutorID
	vs   map[string]sdktypes.Value
}

func (m *executor) ExecutorIDs() []sdktypes.ExecutorID { return m.xids }
func (m *executor) Values() map[string]sdktypes.Value  { return m.vs }

func NewExecutor(caller Caller, xids []sdktypes.ExecutorID, vs map[string]sdktypes.Value) Executor {
	return &executor{Caller: caller, xids: xids, vs: vs}
}
