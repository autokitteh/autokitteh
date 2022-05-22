package locallangrun

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

var (
	ErrWrongState = errors.New("run in current state does not accept operation")
)

type ret struct {
	Err error
	M   map[string]*apivalues.Value
	V   *apivalues.Value
	Sum *apilang.RunSummary
}

type run struct {
	scope string
	id    langrun.RunID
	l     L.Nullable
	runs  *runs

	ctx    context.Context
	cancel func()

	send langrun.SendFunc

	cat    lang.Catalog
	mod    *apiprogram.Module
	vfunc  *apivalues.Value
	kwargs map[string]*apivalues.Value
	args   []*apivalues.Value

	state        *apilang.RunState
	log          []*apilang.RunStateLogRecord
	prints       []string
	cancelReason string
	ret          chan *ret
	lock         sync.RWMutex

	ch chan *apilang.RunState
}

var _ langrun.Run = &run{}

func (r *run) ID() langrun.RunID { return r.id }

func (r *run) summary() *apilang.RunSummary {
	return apilang.NewRunSummary(r.log, r.prints)
}

func (r *run) Summary(context.Context) (*apilang.RunSummary, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.summary(), nil
}

// Unlocked version, when lock scope need to be larger.
// Must be called after lock is acquired.
func (r *run) unlockedUpdate(s *apilang.RunState) {
	now := time.Now()

	r.log = append(r.log, apilang.NewRunStateLogRecord(s, &now))

	prev := r.state

	if s.IsState() {
		r.state = s
	}

	r.send(r.id, now, prev, s)
}

func (r *run) update(s *apilang.RunState) {
	r.lock.Lock()
	r.unlockedUpdate(s)
	r.lock.Unlock()
}

func (r *run) Cancel(_ context.Context, reason string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.state.IsFinal() {
		return ErrWrongState
	}

	r.cancelReason = reason
	r.cancel()

	return nil
}

func (r *run) prn(s string) {
	r.prints = append(r.prints, s)
	r.update(apilang.NewPrintRunUpdateState(s))
}

func (r *run) load(ctx context.Context, path *apiprogram.Path) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	r.l.Debug("load requested", "path", path.String())

	r.lock.Lock()
	r.ret = make(chan *ret)
	r.unlockedUpdate(apilang.NewLoadWaitRunState(path))
	r.lock.Unlock()

	var ret *ret

	select {
	case ret = <-r.ret:
		r.l.Debug("load returned", "ret", ret)
	case <-r.ctx.Done():
		r.l.Debug("runner context error", "err", r.ctx.Err())
		return nil, nil, r.ctx.Err()
	case <-ctx.Done():
		r.l.Debug("load context error", "err", ctx.Err())
		return nil, nil, ctx.Err()
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.ret = nil

	r.unlockedUpdate(apilang.NewLoadReturnedRunUpdate(
		ret.M,
		apiprogram.ImportError(ret.Err),
		ret.Sum,
	))

	if ret.Err != nil {
		return nil, ret.Sum, ret.Err
	}

	return ret.M, ret.Sum, nil
}

func (r *run) call(ctx context.Context, cv *apivalues.Value, kwargs map[string]*apivalues.Value, args []*apivalues.Value, sum *apilang.RunSummary) (*apivalues.Value, error) {
	r.l.Debug("call requested", "call", cv)

	r.lock.Lock()
	r.ret = make(chan *ret)
	r.unlockedUpdate(
		apilang.NewCallWaitRunState(
			cv,
			args,
			kwargs,
			nil, // r.summary(),
		),
	)
	r.lock.Unlock()

	var ret *ret

	select {
	case ret = <-r.ret:
		r.l.Debug("call returned", "ret", ret)
	case <-r.ctx.Done():
		r.l.Debug("context error", "err", r.ctx.Err())
		return nil, r.ctx.Err()
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.ret = nil

	r.unlockedUpdate(apilang.NewCallReturnedRunUpdate(
		ret.V,
		apiprogram.ImportError(ret.Err),
	))

	if ret.Err != nil {
		return nil, ret.Err
	}

	return ret.V, nil
}

func (r *run) ReturnLoad(_ context.Context, m map[string]*apivalues.Value, err error, sum *apilang.RunSummary) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if _, ok := r.state.Get().(*apilang.LoadWaitRunState); !ok {
		return ErrWrongState
	}

	r.l.Debug("returning load", "err", err)

	r.ret <- &ret{M: m, Err: err, Sum: sum}

	return nil
}

func (r *run) ReturnCall(_ context.Context, v *apivalues.Value, err error) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if _, ok := r.state.Get().(*apilang.CallWaitRunState); !ok {
		return ErrWrongState
	}

	r.l.Debug("returning call", "err", err)

	r.ret <- &ret{V: v, Err: err}

	return nil
}

func (r *run) callFunction() {
	env := lang.RunEnv{
		Scope: r.scope,
		Print: r.prn,
		Call:  r.call,
	}

	r.update(apilang.NewRunningRunState())

	_, retv, _, err := langtools.CallFunction(
		r.ctx,
		r.cat,
		&env,
		r.vfunc,
		r.args,
		r.kwargs,
	)

	r.lock.Lock()
	defer r.lock.Unlock()

	if err != nil {
		r.l.Debug("run returned", "err", err)

		if e := (&lang.ErrCanceled{}); errors.As(err, &e) {
			r.unlockedUpdate(apilang.NewCanceledRunState(r.cancelReason, e.CallStack))
			return
		}

		if errors.Is(err, context.Canceled) {
			r.unlockedUpdate(apilang.NewCanceledRunState(r.cancelReason, nil))
			return
		}

		r.unlockedUpdate(apilang.NewErrorRunState(apiprogram.ImportError(err)))
		return
	}

	r.unlockedUpdate(apilang.NewCompletedRunState(
		map[string]*apivalues.Value{
			"return": retv,
		},
	))
}

func (r *run) run() {
	env := lang.RunEnv{
		Scope:    r.scope,
		Predecls: r.kwargs,
		Print:    r.prn,
		Load:     r.load,
		Call:     r.call,
	}

	r.update(apilang.NewRunningRunState())

	_, vs, _, err := langtools.RunModule(
		r.ctx,
		r.cat,
		&env,
		r.mod,
	)

	r.lock.Lock()
	defer r.lock.Unlock()

	if err != nil {
		r.l.Debug("run returned", "err", err)

		if e := (&lang.ErrCanceled{}); errors.As(err, &e) {
			r.unlockedUpdate(apilang.NewCanceledRunState(r.cancelReason, e.CallStack))
			return
		}

		if errors.Is(err, context.Canceled) {
			r.unlockedUpdate(apilang.NewCanceledRunState(r.cancelReason, nil))
			return
		}

		r.unlockedUpdate(apilang.NewErrorRunState(apiprogram.ImportError(err)))
		return
	}

	r.unlockedUpdate(apilang.NewCompletedRunState(vs))
}

func (r *run) Discard(context.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if !r.state.IsDiscardable() {
		return ErrWrongState
	}

	r.runs.discard(r.id)

	return nil
}

func CallFunction(
	l L.L,
	runs *runs,
	id langrun.RunID,
	cat lang.Catalog,
	v *apivalues.Value,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	fn, ok := v.Get().(apivalues.FunctionValue)
	if !ok {
		return nil, errors.New("value is not a function")
	}

	ctx, cancel := context.WithCancel(context.Background())

	if send == nil {
		send = func(langrun.RunID, time.Time, *apilang.RunState, *apilang.RunState) {}
	}

	if cat == nil {
		cat = langtools.DeterministicCatalog
	}

	r := &run{
		scope:  fn.Scope,
		id:     id,
		runs:   runs,
		send:   send,
		cat:    cat,
		ch:     make(chan *apilang.RunState), // no buffering - blocks for listener.
		ctx:    ctx,
		cancel: cancel,
		l:      L.N(l),
		kwargs: kwargs,
		args:   args,
		vfunc:  v,
	}

	go func() {
		defer cancel()

		r.callFunction()
	}()

	return r, nil
}

func RunModule(
	l L.L,
	runs *runs,
	scope string,
	id langrun.RunID,
	cat lang.Catalog,
	mod *apiprogram.Module,
	predecls map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if send == nil {
		send = func(langrun.RunID, time.Time, *apilang.RunState, *apilang.RunState) {}
	}

	if cat == nil {
		cat = langtools.DeterministicCatalog
	}

	r := &run{
		scope:  scope,
		id:     id,
		runs:   runs,
		send:   send,
		cat:    cat,
		mod:    mod,
		kwargs: predecls,
		ch:     make(chan *apilang.RunState), // no buffering - blocks for listener.
		ctx:    ctx,
		cancel: cancel,
		l:      L.N(l),
	}

	go func() {
		defer cancel()

		r.run()
	}()

	return r, nil
}
