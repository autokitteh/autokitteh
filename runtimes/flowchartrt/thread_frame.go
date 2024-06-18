package flowchartrt

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type frame struct {
	node       *threadNode
	states     map[string]sdktypes.Value
	args       map[string]sdktypes.Value
	lastResult sdktypes.Value
}

func (f *frame) withArgs(args map[string]sdktypes.Value) *frame {
	ff := *f
	ff.args = args
	return &ff
}

func (f *frame) getState(k string) map[string]sdktypes.Value {
	if f.states == nil {
		return nil
	}

	return f.states[k].GetStruct().Fields()
}

func (f *frame) setState(k string, v map[string]sdktypes.Value) {
	if v != nil {
		if f.states == nil {
			f.states = make(map[string]sdktypes.Value)
		}

		sym := sdktypes.NewSymbolValue(sdktypes.NewSymbol(k))

		f.states[k] = kittehs.Must1(sdktypes.NewStructValue(sym, v))

		return
	}

	if f.states != nil {
		delete(f.states, k)
	}
}

func (f *frame) updateResult(update func(sdktypes.Value) sdktypes.Value) {
	if f.states == nil {
		f.states = make(map[string]sdktypes.Value)
	}

	result := f.states["results"]

	if !result.IsValid() {
		result = kittehs.Must1(sdktypes.NewDictValue(nil))
	}

	k := sdktypes.NewStringValue(f.node.node.Name)

	curr := kittehs.Must1(result.GetKey(k))

	result = kittehs.Must1(result.SetKey(k, update(curr)))

	f.states["results"] = result
}

func (f *frame) setResult(v sdktypes.Value) {
	f.updateResult(func(sdktypes.Value) sdktypes.Value { return v })
}

func (f *frame) getResult() sdktypes.Value {
	if f.states == nil {
		return sdktypes.InvalidValue
	}

	result := f.states["results"]

	if !result.IsValid() {
		return sdktypes.InvalidValue
	}

	return kittehs.Must1(result.GetKey(sdktypes.NewStringValue(f.node.node.Name)))
}
