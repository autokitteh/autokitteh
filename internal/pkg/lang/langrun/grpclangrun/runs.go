package grpclangrun

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	pblangsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/langsvc"

	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langrun"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type runs struct {
	runClient pblangsvc.LangRunClient
	l         L.Nullable
}

func (r *runs) CallFunction(
	ctx context.Context,
	id langrun.RunID,
	fn *apivalues.Value,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	return CallFunction(ctx, r.l.With("id", id), r.runClient, id, fn, args, kwargs, send)
}

func (r *runs) RunModule(
	ctx context.Context,
	scope string,
	id langrun.RunID,
	mod *apiprogram.Module,
	predecls map[string]*apivalues.Value,
	send langrun.SendFunc,
) (langrun.Run, error) {
	return RunModule(ctx, r.l.With("id", id), r.runClient, scope, id, mod, predecls, send)
}

func (r *runs) Get(ctx context.Context, id langrun.RunID) (langrun.Run, error) {
	_, err := r.runClient.RunGet(ctx, &pblangsvc.RunGetRequest{RunId: string(id)})
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &run{client: r.runClient, id: id, l: L.N(r.l.With("id", id))}, nil
}

func (r *runs) List(ctx context.Context) (map[string]map[langrun.RunID]bool, error) {
	resp, err := r.runClient.ListRuns(ctx, &pblangsvc.ListRunsRequest{})
	if err != nil {
		return nil, err
	}

	m := make(map[string]map[langrun.RunID]bool, len(resp.States))

	for state, pbruns := range resp.States {
		runs := make(map[langrun.RunID]bool, len(pbruns.Runs))

		for _, pbrun := range pbruns.Runs {
			runs[langrun.RunID(pbrun.Id)] = true
		}

		m[state] = runs
	}

	return m, nil
}

func NewRuns(l L.L, client pblangsvc.LangRunClient) langrun.Runs {
	return &runs{l: L.N(l), runClient: client}
}

func NewRunsFromConn(l L.L, conn *grpc.ClientConn) langrun.Runs {
	return NewRuns(l, pblangsvc.NewLangRunClient(conn))
}
