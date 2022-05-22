package langrungrpcsvc

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pblangsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/langsvc"

	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
	L "github.com/autokitteh/autokitteh/pkg/l"

	_ "github.com/autokitteh/autokitteh/internal/pkg/lang/langall"
)

type Svc struct {
	pblangsvc.UnimplementedLangRunServer

	L    L.Nullable
	Runs langrun.Runs
}

var _ pblangsvc.LangRunServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pblangsvc.RegisterLangRunServer(srv, s)

	if gw != nil {
		if err := pblangsvc.RegisterLangRunHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) CallFunction(req *pblangsvc.CallFunctionRequest, srv pblangsvc.LangRun_CallFunctionServer) error {
	l := s.L.With("id", req.RunId)

	l.Debug("CallFunction called")

	if err := req.Validate(); err != nil {
		return status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	args, err := apivalues.ValuesListFromProto(req.Args)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid kwargs: %v", err)
	}

	kwargs, err := apivalues.StringValueMapFromProto(req.Kwargs)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid kwargs: %v", err)
	}

	ch := make(chan *pblangsvc.RunUpdate)

	send := func(id langrun.RunID, t time.Time, prev, next *apilang.RunState) {
		l.Debug("received update", "t", t, "prev", prev.Name(), "next", next.Name())

		if ch == nil {
			l.Debug("channel is closed")
			return
		}

		ch <- &pblangsvc.RunUpdate{
			RunId: string(id),
			T:     timestamppb.New(t),
			Prev:  prev.PB(),
			Next:  next.PB(),
		}

		if next.IsFinal() {
			close(ch)
			ch = nil
		}
	}

	fn, err := apivalues.ValueFromProto(req.F)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid function value: %v", err)
	}

	run, err := s.Runs.CallFunction(
		srv.Context(),
		langrun.RunID(req.RunId),
		fn,
		args,
		kwargs,
		send,
	)

	if err != nil {
		return status.Errorf(codes.Unknown, "run: %v", err)
	}

	l = l.With("id", run.ID())

	for upd := range ch {
		l.Debug("relaying update", "update", upd)

		if err := srv.Send(upd); err != nil {
			return status.Errorf(codes.Unknown, "send: %v", err)
		}
	}

	l.Debug("finished")

	return nil
}

func (s *Svc) Run(req *pblangsvc.RunRequest, srv pblangsvc.LangRun_RunServer) error {
	l := s.L.With("id", req.Id, "scope", req.Scope)

	l.Debug("Run called")

	if err := req.Validate(); err != nil {
		return status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	predecls, err := apivalues.StringValueMapFromProto(req.Predecls)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid predecls: %v", err)
	}

	ch := make(chan *pblangsvc.RunUpdate)

	send := func(id langrun.RunID, t time.Time, prev, next *apilang.RunState) {
		l.Debug("received update", "t", t, "prev", prev.Name(), "next", next.Name())

		if ch == nil {
			l.Debug("channel is closed")
			return
		}

		ch <- &pblangsvc.RunUpdate{
			RunId: string(id),
			T:     timestamppb.New(t),
			Prev:  prev.PB(),
			Next:  next.PB(),
		}

		if next.IsFinal() {
			l.Debug("final state, closing channel", "next", next)
			close(ch)
			ch = nil
		}
	}

	mod, err := apiprogram.ModuleFromProto(req.Module)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid module: %v", err)
	}

	run, err := s.Runs.RunModule(
		srv.Context(),
		req.Scope,
		langrun.RunID(req.Id),
		mod,
		predecls,
		send,
	)

	if err != nil {
		return status.Errorf(codes.Unknown, "run: %v", err)
	}

	l = l.With("id", run.ID())

	for upd := range ch {
		l.Debug("relaying update", "update", upd)

		if err := srv.Send(upd); err != nil {
			return status.Errorf(codes.Unknown, "send: %v", err)
		}
	}

	l.Debug("finished")

	return nil
}

func (s *Svc) RunCallReturn(ctx context.Context, req *pblangsvc.RunCallReturnRequest) (*pblangsvc.RunCallReturnResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	retval, err := apivalues.ValueFromProto(req.Retval)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "values: %v", err)
	}

	callErr, err := apiprogram.ErrorFromProto(req.Error)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error: %v", err)
	}

	id := langrun.RunID(req.RunId)

	run, err := s.Runs.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	if run == nil {
		return nil, status.Errorf(codes.NotFound, "get")
	}

	// Error cast as error might fail nil check even if nil.
	var retCallErr error
	if callErr != nil {
		retCallErr = callErr
	}

	if err := run.ReturnCall(ctx, retval, retCallErr); err != nil {
		return nil, status.Errorf(codes.Unknown, "return: %v", err)
	}

	return &pblangsvc.RunCallReturnResponse{}, nil
}

func (s *Svc) RunLoadReturn(ctx context.Context, req *pblangsvc.RunLoadReturnRequest) (*pblangsvc.RunLoadReturnResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	vals, err := apivalues.StringValueMapFromProto(req.Values)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "values: %v", err)
	}

	loadErr, err := apiprogram.GOErrorFromProto(req.Error)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error: %v", err)
	}

	sum, err := apilang.RunSummaryFromProto(req.RunSummary)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "run summary: %v", err)
	}

	id := langrun.RunID(req.RunId)

	run, err := s.Runs.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	if run == nil {
		return nil, status.Errorf(codes.NotFound, "get")
	}

	if err := run.ReturnLoad(ctx, vals, loadErr, sum); err != nil {
		return nil, status.Errorf(codes.Unknown, "return: %v", err)
	}

	return &pblangsvc.RunLoadReturnResponse{}, nil
}

func (s *Svc) RunCancel(ctx context.Context, req *pblangsvc.RunCancelRequest) (*pblangsvc.RunCancelResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	id := langrun.RunID(req.RunId)

	run, err := s.Runs.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	if run == nil {
		return nil, status.Errorf(codes.NotFound, "get")
	}

	if err := run.Cancel(ctx, req.Reason); err != nil {
		return nil, status.Errorf(codes.Unknown, "cancel: %v", err)
	}

	return &pblangsvc.RunCancelResponse{}, nil
}

func (s *Svc) RunGet(ctx context.Context, req *pblangsvc.RunGetRequest) (*pblangsvc.RunGetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	id := langrun.RunID(req.RunId)

	run, err := s.Runs.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	if run == nil {
		return nil, status.Errorf(codes.NotFound, "get")
	}

	if req.GetSummary {
		sum, err := run.Summary(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "summary: %v", err)
		}

		return &pblangsvc.RunGetResponse{Summary: sum.PB()}, nil
	}

	return &pblangsvc.RunGetResponse{}, nil
}

func (s *Svc) ListRuns(ctx context.Context, req *pblangsvc.ListRunsRequest) (*pblangsvc.ListRunsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	l, err := s.Runs.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "list: %v", err)
	}

	pb := make(map[string]*pblangsvc.ListRuns)

	for state, runs := range l {
		pbruns := make([]*pblangsvc.ListRun, 0, len(runs))
		for run, v := range runs {
			if !v {
				continue
			}

			pbruns = append(pbruns, &pblangsvc.ListRun{Id: string(run)})
		}

		pb[state] = &pblangsvc.ListRuns{Runs: pbruns}
	}

	return &pblangsvc.ListRunsResponse{States: pb}, nil
}

func (s *Svc) RunDiscard(ctx context.Context, req *pblangsvc.RunDiscardRequest) (*pblangsvc.RunDiscardResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	id := langrun.RunID(req.Id)

	run, err := s.Runs.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	if run == nil {
		return nil, status.Errorf(codes.NotFound, "get")
	}

	if err := run.Discard(ctx); err != nil {
		return nil, status.Errorf(codes.Unknown, "discard: %v", err)
	}

	return &pblangsvc.RunDiscardResponse{}, nil
}
