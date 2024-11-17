package sessiondata

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Data struct {
	Session     sdktypes.Session        `json:"session"`
	Vars        []sdktypes.Var          `json:"vars"`
	Build       sdktypes.Build          `json:"build"`
	BuildFile   *sdkbuildfile.BuildFile `json:"build_file"`
	Triggers    []sdktypes.Trigger      `json:"mappings"`
	Connections []sdktypes.Connection   `json:"connections"`
}

type ConnInfo struct {
	Config          map[string]string `json:"config"`
	IntegrationName string            `json:"integration_name"`
}

func retrieve[I sdktypes.ID, R sdktypes.Object](ctx context.Context, id I, f func(context.Context, I) (R, error)) (R, error) {
	var invalid R

	r, err := sdkerrors.IgnoreNotFoundErr(f(ctx, id))

	if err != nil {
		return invalid, fmt.Errorf("get %q: %w", id, err)
	} else if !r.IsValid() {
		return invalid, fmt.Errorf("%q not found", id)
	}

	return r, nil
}

// TODO(ENG-205): Limit max size.
func downloadBuild(ctx context.Context, buildID sdktypes.BuildID, builds sdkservices.Builds) ([]byte, error) {
	r, err := builds.Download(ctx, buildID)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	return io.ReadAll(r)
}

func Get(ctx context.Context, svcs *sessionsvcs.Svcs, session sdktypes.Session) (*Data, error) {
	var err error

	data := Data{Session: session}

	// TODO(ENG-207): Consider doing all retrievals using one big happy join.

	if pid := session.ProjectID(); pid.IsValid() {
		cfilter := sdkservices.ListConnectionsFilter{ProjectID: pid}
		if data.Connections, err = svcs.Connections.List(ctx, cfilter); err != nil {
			return nil, fmt.Errorf("connections.list: %w", err)
		}

		tfilter := sdkservices.ListTriggersFilter{ProjectID: pid}
		if data.Triggers, err = svcs.Triggers.List(ctx, tfilter); err != nil {
			return nil, fmt.Errorf("triggers.list(%v): %w", pid, err)
		}

		if data.Vars, err = svcs.Vars.Get(ctx, sdktypes.NewVarScopeID(pid)); err != nil {
			return nil, fmt.Errorf("get vars: %w", err)
		}
	}

	buildID := data.Session.BuildID()

	if data.Build, err = retrieve(ctx, buildID, svcs.Builds.Get); err != nil {
		return nil, err
	}

	buildData, err := downloadBuild(ctx, buildID, svcs.Builds)
	if err != nil {
		return nil, fmt.Errorf("download build: %w", err)
	}

	if data.BuildFile, err = sdkbuildfile.Read(bytes.NewBuffer(buildData)); err != nil {
		return nil, fmt.Errorf("corrupted build file: %w", err)
	}

	return &data, nil
}
