package langgrpc

import (
	"context"
	"fmt"
	"time"

	pblangsvc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/langsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun/grpclangrun"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apilang"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

var ErrUnsupported = fmt.Errorf("unsupported")

type grpcLang struct {
	name      string
	client    pblangsvc.LangClient
	runClient pblangsvc.LangRunClient
	l         L.Nullable
}

func MustNew(l L.L, name string, client pblangsvc.LangClient, runClient pblangsvc.LangRunClient) lang.Lang {
	gl, err := New(l, name, client, runClient)
	if err != nil {
		panic(err)
	}
	return gl
}

func New(l L.L, name string, client pblangsvc.LangClient, runClient pblangsvc.LangRunClient) (lang.Lang, error) {
	return &grpcLang{l: L.N(l), name: name, client: client, runClient: runClient}, nil
}

func (gl *grpcLang) IsCompilerVersionSupported(ctx context.Context, v string) (bool, error) {
	if gl.client == nil {
		return false, ErrUnsupported
	}

	resp, err := gl.client.IsCompilerVersionSupported(
		ctx,
		&pblangsvc.IsCompilerVersionSupportedRequest{Lang: gl.name, Ver: v},
	)
	if err != nil {
		return false, err
	}

	return resp.Supported, nil
}

func (gl *grpcLang) CompileModule(
	ctx context.Context,
	path *apiprogram.Path,
	src []byte,
	predecls []string,
) (*apiprogram.Module, error) {
	if gl.client == nil {
		return nil, ErrUnsupported
	}

	resp, err := gl.client.CompileModule(
		ctx,
		&pblangsvc.CompileModuleRequest{
			Lang:     gl.name,
			Predecls: predecls,
			Path:     path.PB(),
			Src:      src,
		},
	)
	if err != nil {
		if perr := apiprogram.ErrorFromGRPCError(err); perr != nil {
			return nil, perr
		}

		return nil, err
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	mod, err := apiprogram.ModuleFromProto(resp.Module)
	if err != nil {
		return nil, fmt.Errorf("invalid module proto: %w", err)
	}

	return mod, nil
}

func (gl *grpcLang) GetModuleDependencies(ctx context.Context, mod *apiprogram.Module) ([]*apiprogram.Path, error) {
	if gl.client == nil {
		return nil, ErrUnsupported
	}

	resp, err := gl.client.GetModuleDependencies(
		ctx,
		&pblangsvc.GetModuleDependenciesRequest{Module: mod.PB()},
	)
	if err != nil {
		return nil, err
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	paths := make([]*apiprogram.Path, len(resp.Deps.Ready))
	for i, pbp := range resp.Deps.Ready {
		if paths[i], err = apiprogram.PathFromProto(pbp); err != nil {
			return nil, fmt.Errorf("invalid path %d: %q", i, err)
		}
	}

	return paths, nil
}

func (gl *grpcLang) run(
	ctx context.Context,
	env *lang.RunEnv,
	runFn func(context.Context, langrun.RunID, langrun.SendFunc) (langrun.Run, error),
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	if gl.runClient == nil {
		return nil, nil, ErrUnsupported
	}

	env = env.WithStubs()

	id := langrun.NewRunID()

	ch := make(chan *apilang.RunState)

	l := gl.l.With("id", id)

	sum := apilang.NewRunSummary(nil, nil)

	send := func(_ langrun.RunID, t time.Time, prev, next *apilang.RunState) {
		l := l.With("t", t, "prev", prev.Name(), "next", next.Name())

		l.Debug("received update")

		if ch == nil {
			l.Warn("channel is already closed")
			return
		}

		if next != nil {
			sum.Add(apilang.NewRunStateLogRecord(next, &t))
		}

		ch <- next
		if next.IsFinal() {
			close(ch)
			ch = nil
		}
	}

	l.Debug("initiating run")

	runCtx := context.Background()

	run, err := runFn(runCtx, id, send)
	if err != nil {
		return nil, nil, fmt.Errorf("run: %w", err)
	}

	// this cancel might be used when the context is canceled, and a new context is set
	// to allow operations to be finished.
	cancel := func() {}

	l.Debug("run started")

	for {
		l.Debug("waiting")

		select {
		case state := <-ch:
			l.Debug("received updated", "state", state.Name())

			switch s := state.Get().(type) {
			case *apilang.LoadWaitRunState:
				l.Debug("loading")

				vs, sum, err := env.Load(ctx, s.Path())

				l.Debug("returning load", "err", err)

				if err := run.ReturnLoad(ctx, vs, err, sum); err != nil {
					cancel()
					return nil, sum, fmt.Errorf("return load: %w", err)
				}

			case *apilang.CallWaitRunState:
				l.Debug("calling")

				v, err := env.Call(ctx, s.CallValue(), s.Kws(), s.Args(), s.RunSummary())

				l.Debug("returning call", "err", err)

				if err := run.ReturnCall(ctx, v, err); err != nil {
					cancel()
					return nil, sum, fmt.Errorf("return call: %w", err)
				}

			case *apilang.PrintRunUpdate:
				env.Print(s.Text())
			case *apilang.CompletedRunState:
				cancel()
				return s.Values(), sum, nil
			case *apilang.ErrorRunState:
				cancel()
				return nil, sum, s.Error()
			case *apilang.CanceledRunState:
				cancel()
				return nil, sum, &lang.ErrCanceled{CallStack: s.CallStack()}
			case *apilang.ClientErrorRunState:
				cancel()
				return nil, sum, s
			default:
				// ignore
			}
		case <-ctx.Done():
			l.Debug("context is done", "err", ctx.Err())

			// the original context is done. make time limitted contexts for the rest of
			// the operations

			var cancelCtx context.Context

			cancelCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

			if cerr := run.Cancel(cancelCtx, "context canceled"); cerr != nil {
				cancel()
				return nil, sum, fmt.Errorf("cancel error: %w", cerr)
			}

			cancel()

			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)

			// go on to next loop to get a chance to get that cancel update.
		}
	}
}

func (gl *grpcLang) RunModule(
	ctx context.Context,
	env *lang.RunEnv,
	mod *apiprogram.Module, // mod must have compiled_code populated.
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	return gl.run(
		ctx,
		env,
		func(ctx context.Context, id langrun.RunID, send langrun.SendFunc) (langrun.Run, error) {
			return grpclangrun.RunModule(ctx, gl.l.With("id", id), gl.runClient, env.Scope, id, mod, env.Predecls, send)
		},
	)
}

func (gl *grpcLang) CallFunction(
	ctx context.Context,
	env *lang.RunEnv,
	fn *apivalues.Value,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, *apilang.RunSummary, error) {
	m, sum, err := gl.run(
		ctx,
		env,
		func(ctx context.Context, id langrun.RunID, send langrun.SendFunc) (langrun.Run, error) {
			return grpclangrun.CallFunction(ctx, gl.l.With("id", id), gl.runClient, id, fn, args, kwargs, send)
		},
	)

	var ret *apivalues.Value

	if m != nil {
		ret = m["return"]
	}

	return ret, sum, err
}
