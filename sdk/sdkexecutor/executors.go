package sdkexecutor

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Executors struct {
	values  map[string]map[string]sdktypes.Value // scope -> {values}; scope is usually module name.
	callers map[string]Caller                    // xid -> caller
}

func (ms *Executors) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if !v.IsFunction() {
		return sdktypes.InvalidValue, sdkerrors.ErrInvalidArgument{}
	}

	xid := v.GetFunction().ExecutorID()
	c := ms.GetCaller(xid)
	if c == nil {
		return sdktypes.InvalidValue, fmt.Errorf("executor not found: %w", sdkerrors.ErrNotFound)
	}

	return c.Call(ctx, v, args, kwargs)
}

func (ms *Executors) AddExecutor(name string, x Executor) error {
	if err := ms.AddCaller(x.ExecutorID(), x); err != nil {
		return fmt.Errorf("add caller: %w", err)
	}

	if err := ms.AddValues(name, x.Values()); err != nil {
		return fmt.Errorf("add values: %w", err)
	}

	return nil
}

func (ms *Executors) AddCaller(xid sdktypes.ExecutorID, c Caller) error {
	if ms.callers == nil {
		ms.callers = make(map[string]Caller)
	}

	if _, ok := ms.callers[xid.String()]; ok {
		return sdkerrors.ErrConflict
	}

	ms.callers[xid.String()] = c

	return nil
}

func (ms *Executors) AddValues(scope string, vs map[string]sdktypes.Value) error {
	if ms.values == nil {
		ms.values = make(map[string]map[string]sdktypes.Value)
	}

	if _, ok := ms.values[scope]; ok {
		return sdkerrors.ErrConflict
	}

	ms.values[scope] = vs

	return nil
}

func (ms *Executors) ValuesForScopePrefix(prefix string) map[string]map[string]sdktypes.Value {
	return kittehs.FilterMapKeys(ms.values, func(n string) bool {
		return strings.HasPrefix(n, prefix)
	})
}

func (ms *Executors) GetCaller(xid sdktypes.ExecutorID) Caller {
	return ms.callers[xid.String()]
}

func (ms *Executors) GetValues(scope string) map[string]sdktypes.Value {
	return ms.values[scope]
}
