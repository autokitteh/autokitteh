package locallangrun

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
	L "github.com/autokitteh/L"
)

type RunModuleFunc func(L.L, string, langrun.RunID, *apiprogram.Module, map[string]*apivalues.Value, langrun.SendFunc) (langrun.Run, error)

type CallFunctionFunc func(L.L, langrun.RunID, *apivalues.Value, []*apivalues.Value, map[string]*apivalues.Value, langrun.SendFunc) (langrun.Run, error)

type runs struct {
	runModule    RunModuleFunc
	callFunction CallFunctionFunc

	runs        map[langrun.RunID]langrun.Run // TODO: cleanup? lru?
	runsByState map[string]map[langrun.RunID]bool
	lock        sync.RWMutex

	l L.Nullable
}

func (r *runs) Get(_ context.Context, id langrun.RunID) (langrun.Run, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.runs[id], nil
}

func (r *runs) List(context.Context) (map[string]map[langrun.RunID]bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	m := make(map[string]map[langrun.RunID]bool)
	for s, ids := range r.runsByState {
		rs := make(map[langrun.RunID]bool)
		for id, v := range ids {
			rs[id] = v
		}
		m[s] = rs
	}

	return m, nil
}

func (r *runs) discard(id langrun.RunID) {
	l := r.l.With("id", id)

	l.Debug("discarding")

	r.lock.Lock()
	defer r.lock.Unlock()

	grun, ok := r.runs[id]
	if !ok {
		l.Warn("run not found")
		return
	}

	lrun, ok := grun.(*run)
	if !ok {
		l.Error("run is not a local run")
		return
	}

	delete(r.runsByState[lrun.state.Name()], id)
	delete(r.runs, id)
}

func (r *runs) wrapSend(send langrun.SendFunc) langrun.SendFunc {
	return func(id langrun.RunID, t time.Time, prev *apilang.RunState, state *apilang.RunState) {
		// keep track of state changes.

		if state.IsState() {
			r.lock.Lock()

			if prev != nil {
				n := prev.Name()

				delete(r.runsByState[n], id)
			}

			n := state.Name()

			rs := r.runsByState[n]
			if rs == nil {
				rs = make(map[langrun.RunID]bool)
			}

			rs[id] = true
			r.runsByState[n] = rs

			r.lock.Unlock()
		}

		send(id, t, prev, state)
	}
}

func (r *runs) RunModule(
	ctx context.Context,
	scope string,
	id langrun.RunID,
	mod *apiprogram.Module,
	predecls map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	r.lock.Lock()
	if _, found := r.runs[id]; found {
		return nil, fmt.Errorf("run %v already exists", id)
	}
	// claim the id
	r.runs[id] = nil
	r.lock.Unlock()

	run, err := r.runModule(r.l.With("id", id), scope, id, mod, predecls, r.wrapSend(send))
	if err != nil {
		return nil, err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.runs[id] = run

	return run, nil
}

func (r *runs) CallFunction(
	ctx context.Context,
	id langrun.RunID,
	fn *apivalues.Value,
	args []*apivalues.Value,
	kws map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	r.lock.Lock()
	if _, found := r.runs[id]; found {
		r.lock.Unlock()
		return nil, fmt.Errorf("run %v already exists", id)
	}
	// claim the id
	r.runs[id] = nil
	r.lock.Unlock()

	run, err := r.callFunction(r.l.With("id", id), id, fn, args, kws, r.wrapSend(send))
	if err != nil {
		return nil, err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.runs[id] = run

	return run, nil
}

func NewRuns(l L.L, cat lang.Catalog, runModule RunModuleFunc, callFunction CallFunctionFunc) langrun.Runs {
	runs := &runs{
		runs:        make(map[langrun.RunID]langrun.Run),
		runsByState: make(map[string]map[langrun.RunID]bool),
		l:           L.N(l),
	}

	if runModule == nil {
		runModule = func(l L.L, scope string, id langrun.RunID, mod *apiprogram.Module, predecls map[string]*apivalues.Value, send langrun.SendFunc) (langrun.Run, error) {
			return RunModule(l, runs, scope, id, cat, mod, predecls, send)
		}
	}

	if callFunction == nil {
		callFunction = func(l L.L, id langrun.RunID, v *apivalues.Value, args []*apivalues.Value, kws map[string]*apivalues.Value, send langrun.SendFunc) (langrun.Run, error) {
			return CallFunction(l, runs, id, cat, v, args, kws, send)
		}
	}

	runs.runModule = runModule
	runs.callFunction = callFunction

	return runs
}
