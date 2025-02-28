package tempokitteh

import (
	"context"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var vw = sdktypes.ValueWrapper{SafeForJSON: true}

type tk struct {
	l           *zap.Logger
	runtimes    sdkservices.Runtimes
	calls       sessioncalls.Calls
	worker      Worker
	entrypoints map[string]sdktypes.CodeLocation
	build       *sdkbuildfile.BuildFile
}

func Register(ctx context.Context, l *zap.Logger, w Worker, runtimes sdkservices.Runtimes, calls sessioncalls.Calls, build *sdkbuildfile.BuildFile) error {
	tk := &tk{
		runtimes:    runtimes,
		calls:       calls,
		worker:      w,
		l:           l,
		build:       build,
		entrypoints: make(map[string]sdktypes.CodeLocation),
	}

	for _, rt := range build.Runtimes {
		for _, e := range rt.Artifact.Exports() {
			if !e.Location().IsValid() || !e.Symbol().IsValid() {
				continue
			}

			name := e.Symbol().String()

			tk.entrypoints[name] = e.Location()

			l.Info("registering workflow", zap.String("name", name), zap.String("location", e.Location().String()))

			w.worker().RegisterWorkflowWithOptions(tk.workflow, workflow.RegisterOptions{Name: name})
		}
	}

	return nil
}
