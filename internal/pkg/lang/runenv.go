package lang

import (
	"context"
	"errors"

	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

type PrintFunc func(string)
type LoadFunc func(context.Context, *apiprogram.Path) (map[string]*apivalues.Value, *apilang.RunSummary, error)
type CallFunc func(context.Context, *apivalues.Value, map[string]*apivalues.Value, []*apivalues.Value, *apilang.RunSummary) (*apivalues.Value, error)

type RunEnv struct {
	Scope    string
	Predecls map[string]*apivalues.Value

	// TODO: need to limit size of what can passed in this funcs to prevet DOS.
	Print PrintFunc
	Load  LoadFunc // not relevant for function calls.
	Call  CallFunc
}

func (e *RunEnv) Clone() *RunEnv {
	ee := *e

	for k, v := range e.Predecls {
		ee.Predecls[k] = v.Clone()
	}

	return &ee
}

func (e *RunEnv) WithStubs() *RunEnv {
	if e == nil {
		return EmptyRunEnv.Clone()
	}

	ee := e.Clone()

	if ee.Load == nil {
		ee.Load = EmptyRunEnv.Load
	}

	if ee.Print == nil {
		ee.Print = EmptyRunEnv.Print
	}

	if ee.Call == nil {
		ee.Call = EmptyRunEnv.Call
	}

	return ee
}

var EmptyRunEnv = RunEnv{
	Print: func(string) {},
	Load: func(context.Context, *apiprogram.Path) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
		return nil, nil, errors.New("load is not supported")
	},
	Call: func(context.Context, *apivalues.Value, map[string]*apivalues.Value, []*apivalues.Value, *apilang.RunSummary) (*apivalues.Value, error) {
		return nil, errors.New("call is not supported")
	},
}
