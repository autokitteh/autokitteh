package values

import (
	"context"
	"errors"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const moduleName = "values"

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("values"))

func init() { modules.Register(ExecutorID) }

type module struct {
	pid sdktypes.ProjectID
	db  db.DB
}

func New(pid sdktypes.ProjectID, db db.DB) sdkexecutor.Executor {
	m := &module{pid: pid, db: db}

	return fixtures.NewBuiltinExecutor(
		ExecutorID,

		sdkmodule.ExportFunction("get", m.get),
		sdkmodule.ExportFunction("set", m.set),
	)
}

func (m *module) get(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var key string

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &key); err != nil {
		return sdktypes.InvalidValue, err
	}

	v, err := m.db.GetValue(ctx, m.pid, key)
	if err != nil && errors.Is(err, sdkerrors.ErrNotFound) {
		return sdktypes.Nothing, nil
	}

	return v, err
}

func (m *module) set(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		key string
		val sdktypes.Value
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &key, "value", &val); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.InvalidValue, m.db.SetValue(ctx, m.pid, key, val)
}

// curr: current stored value.
// ret: value returned from operation.
// next: new value to store instead of current.
type mutatorFn func(curr sdktypes.Value) (ret sdktypes.Value, next sdktypes.Value, err error)

func (m *module) mutate(ctx context.Context, key string, f mutatorFn) (sdktypes.Value, error) {
	var ret sdktypes.Value

	err := m.db.Transaction(ctx, func(tx db.DB) error {
		curr, err := tx.GetValue(ctx, m.pid, key)
		if err != nil {
			if !errors.Is(err, sdkerrors.ErrNotFound) {
				return err
			}

			curr = sdktypes.Nothing
		}

		var next sdktypes.Value

		if ret, next, err = f(curr); err != nil {
			return err
		}

		return m.db.SetValue(ctx, m.pid, key, next)
	})
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return ret, nil
}
