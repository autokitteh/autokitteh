package langrun

import (
	"context"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apilang"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
)

type Run interface {
	ID() RunID
	Cancel(context.Context, string) error
	Summary(context.Context) (*apilang.RunSummary, error)
	ReturnLoad(context.Context, map[string]*apivalues.Value, error, *apilang.RunSummary) error
	ReturnCall(context.Context, *apivalues.Value, error) error
	Discard(context.Context) error
}
