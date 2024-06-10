package integrations

import (
	"context"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type test struct{}

var integrationID = sdktypes.NewIntegrationIDFromName("test")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "test",
	DisplayName:   "Test",
	Description:   "Test integration",
}))

func newTestIntegration() sdkservices.Integration {
	var i test

	return sdkintegrations.NewIntegration(desc, sdkmodule.New(
		sdkmodule.ExportFunction(
			"freeze",
			i.freeze,
			sdkmodule.WithArgs("duration?", "allow_cancel?"),
		),
	))
}

func (i test) freeze(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		duration    time.Duration
		allowCancel bool
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "duration?", &duration, "allow_cancel?", &allowCancel); err != nil {
		return sdktypes.InvalidValue, err
	}

	var done <-chan struct{}
	if allowCancel {
		done = ctx.Done()
	}

	var tmo <-chan time.Time
	if duration > 0 {
		tmo = time.After(duration)
	}

	select {
	case <-done:
		return sdktypes.InvalidValue, ctx.Err()
	case <-tmo:
		return sdktypes.Nothing, nil
	}
}
