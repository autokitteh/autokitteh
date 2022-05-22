package grpclangrun

import (
	"context"
	"fmt"
	"time"

	pblangsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/langsvc"

	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
	L "github.com/autokitteh/L"
)

type run struct {
	id     langrun.RunID
	client pblangsvc.LangRunClient
	l      L.Nullable
}

var _ langrun.Run = &run{}

func (r *run) ID() langrun.RunID { return r.id }

func (r *run) Cancel(ctx context.Context, reason string) error {
	_, err := r.client.RunCancel(ctx, &pblangsvc.RunCancelRequest{RunId: string(r.id), Reason: reason})
	return err
}

func (r *run) Discard(ctx context.Context) error {
	_, err := r.client.RunDiscard(ctx, &pblangsvc.RunDiscardRequest{Id: string(r.id)})
	return err
}

func (r *run) ReturnLoad(ctx context.Context, vs map[string]*apivalues.Value, err error, sum *apilang.RunSummary) error {
	_, callErr := r.client.RunLoadReturn(ctx, &pblangsvc.RunLoadReturnRequest{
		RunId:      string(r.id),
		Error:      apiprogram.ImportError(err).PB(),
		Values:     apivalues.StringValueMapToProto(vs),
		RunSummary: sum.PB(),
	})
	return callErr
}

func (r *run) ReturnCall(ctx context.Context, v *apivalues.Value, err error) error {
	_, callErr := r.client.RunCallReturn(ctx, &pblangsvc.RunCallReturnRequest{
		RunId:  string(r.id),
		Error:  apiprogram.ImportError(err).PB(),
		Retval: v.PB(),
	})
	return callErr
}

func (r *run) Summary(ctx context.Context) (*apilang.RunSummary, error) {
	resp, err := r.client.RunGet(ctx, &pblangsvc.RunGetRequest{RunId: string(r.id), GetSummary: true})
	if err != nil {
		return nil, err
	}

	if err := resp.Validate(); err != nil {
		return nil, err
	}

	return apilang.RunSummaryFromProto(resp.Summary)
}

func (r *run) run(rmc pblangsvc.LangRun_RunClient, send langrun.SendFunc) {
	l := r.l

	clientError := func(f string, vs ...interface{}) {
		send(r.id, time.Now(), nil, apilang.NewClientErrorRunState(fmt.Errorf(f, vs...)))
	}

	var state *apilang.RunState

	for !state.IsFinal() {
		l.Debug("waiting")

		upd, err := rmc.Recv()
		if err != nil {
			clientError("recv: %w", err)
			return
		}

		l.Debug("received update", "upd", upd)

		if err := upd.Validate(); err != nil {
			clientError("validate: %w", err)
			return
		}

		prev, err := apilang.RunStateFromProto(upd.Prev)
		if err != nil {
			clientError("prev decode: %w", err)
			return
		}

		next, err := apilang.RunStateFromProto(upd.Next)
		if err != nil {
			clientError("next decode: %w", err)
			return
		}

		state = next

		l.Debug("relaying", "t", upd.T.AsTime(), "prev", prev.Name(), "next", next.Name())

		send(r.id, upd.T.AsTime(), prev, next)
	}
}

func CallFunction(
	ctx context.Context,
	l L.L,
	client pblangsvc.LangRunClient,
	id langrun.RunID,
	f *apivalues.Value,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	run := &run{client: client, id: id, l: L.N(l)}

	req := pblangsvc.CallFunctionRequest{
		RunId:  string(id),
		F:      f.PB(),
		Args:   apivalues.ValuesListToProto(args),
		Kwargs: apivalues.StringValueMapToProto(kwargs),
	}

	run.l.Debug("calling server")

	rmc, err := client.CallFunction(ctx, &req)
	if err != nil {
		return nil, err
	}

	go run.run(rmc, send)

	return run, nil
}

func RunModule(
	ctx context.Context,
	l L.L,
	client pblangsvc.LangRunClient,
	scope string,
	id langrun.RunID,
	mod *apiprogram.Module,
	predecls map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	run := &run{client: client, id: id, l: L.N(l)}

	req := pblangsvc.RunRequest{
		Scope:    scope,
		Id:       string(id),
		Module:   mod.PB(),
		Predecls: apivalues.StringValueMapToProto(predecls),
	}

	run.l.Debug("calling server")

	rmc, err := client.Run(ctx, &req)
	if err != nil {
		return nil, err
	}

	go run.run(rmc, send)

	return run, nil
}
