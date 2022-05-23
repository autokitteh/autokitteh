package langrun

import (
	"context"
	"time"

	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apivalues"
)

type SendFunc func(id RunID, t time.Time, from *apilang.RunState, to *apilang.RunState)

type Runs interface {
	RunModule(
		_ context.Context,
		scope string,
		_ RunID,
		_ *apiprogram.Module,
		_ map[string]*apivalues.Value,
		_ SendFunc,
	) (Run, error)

	CallFunction(
		_ context.Context,
		_ RunID,
		fn *apivalues.Value,
		args []*apivalues.Value,
		kws map[string]*apivalues.Value,
		_ SendFunc,
	) (Run, error)

	Get(context.Context, RunID) (Run, error)

	List(context.Context) (map[string]map[RunID]bool, error)
}
