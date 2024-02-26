package runtime

import (
	"bytes"
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type run struct {
	xid sdktypes.ExecutorID
	vs  map[string]sdktypes.Value
	cbs *sdkservices.RunCallbacks
}

func (r *run) ID() sdktypes.RunID                { return r.xid.ToRunID() }
func (r *run) ExecutorID() sdktypes.ExecutorID   { return r.xid }
func (r *run) Values() map[string]sdktypes.Value { return r.vs }
func (r *run) Close()                            {}

func (r *run) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// you can call other functions here only if this was run from a workflow (only from an entrypoint).

	data := sdktypes.GetFunctionValueData(v)

	if bytes.Compare(data, []byte("miki-data")) == 0 {
		// yay!
		return sdktypes.NewStringValue("miki"), nil
	}

	r.cbs.Call(
		ctx,
		r.xid.ToRunID(),
		sdktypes.NewFunctionValue(r.xid, "kiki", []byte("kiki-data"), nil, nil),
		nil,
		nil,
	)

	return nil, nil
}

func Run(
	ctx context.Context,
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
	givenValues map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	xid := sdktypes.NewExecutorID(runID)

	vs := map[string]sdktypes.Value{
		"miki": sdktypes.NewFunctionValue(
			xid,
			"miki",
			[]byte("miki-data"),
			nil,
			nil,
		),
	}

	return &run{xid: xid, vs: vs, cbs: cbs}, nil
}
