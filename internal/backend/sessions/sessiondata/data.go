package sessiondata

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Data struct {
	SessionID   sdktypes.SessionID      `json:"session_id"`
	ProjectID   sdktypes.ProjectID      `json:"project_id"`
	Session     sdktypes.Session        `json:"session"`
	Deployment  sdktypes.Deployment     `json:"deployment"`
	Env         sdktypes.Env            `json:"env"`
	EnvVars     []sdktypes.EnvVar       `json:"env_vars"`
	Build       sdktypes.Build          `json:"build"`
	BuildFile   *sdkbuildfile.BuildFile `json:"build_file"`
	Triggers    []sdktypes.Trigger      `json:"mappings"`
	Connections []sdktypes.Connection   `json:"connections"`
}

func retrieve[I sdktypes.ID, R sdktypes.Object](ctx context.Context, z *zap.Logger, id I, f func(context.Context, I) (R, error)) (R, error) {
	var invalid R

	r, err := f(ctx, id)

	if err != nil {
		z.DPanic("get", zap.Error(err), zap.String("id", id.String()))
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

// Get session related data using local activities in order not to expose data to Temporal.
func Get(ctx context.Context, z *zap.Logger, svcs *sessionsvcs.Svcs, sessionID sdktypes.SessionID) (*Data, error) {
	var err error

	data := Data{
		SessionID: sessionID,
	}

	// TODO(ENG-207): Consider doing all retrievals using one big happy join.

	if data.Session, err = retrieve(ctx, z, sessionID, svcs.DB.GetSession); err != nil {
		return nil, err
	}

	if data.Deployment, err = retrieve(ctx, z, data.Session.DeploymentID(), svcs.Deployments.Get); err != nil {
		return nil, err
	}

	envID := data.Deployment.EnvID()

	if data.Env, err = retrieve(ctx, z, envID, svcs.Envs.GetByID); err != nil {
		return nil, err
	}

	if data.ProjectID = data.Env.ProjectID(); !data.ProjectID.IsValid() {
		return nil, fmt.Errorf("sessions can only run on projects")
	}

	if data.Connections, err = svcs.Connections.List(ctx, sdkservices.ListConnectionsFilter{ProjectID: data.ProjectID}); err != nil {
		return nil, fmt.Errorf("connections.list: %w", err)
	}

	for _, conn := range data.Connections {
		ts, err := svcs.Triggers.List(ctx, sdkservices.ListTriggersFilter{ConnectionID: conn.ID()})
		if err != nil {
			return nil, fmt.Errorf("triggers.list(%v): %w", conn.ID(), err)
		}

		ts = kittehs.Filter(ts, func(t sdktypes.Trigger) bool {
			triggerEnvID := t.EnvID()
			return !triggerEnvID.IsValid() || triggerEnvID == envID
		})

		data.Triggers = append(data.Triggers, ts...)
	}

	// TODO: merge mappings?

	if data.EnvVars, err = svcs.Envs.GetVars(ctx, nil, envID); err != nil {
		return nil, fmt.Errorf("get vars: %w", err)
	}

	buildID := data.Deployment.BuildID()

	if data.Build, err = retrieve(ctx, z, buildID, svcs.Builds.Get); err != nil {
		return nil, err
	}

	buildData, err := downloadBuild(ctx, buildID, svcs.Builds)
	if err != nil {
		z.Panic("download build data", zap.Error(err), zap.String("build_id", buildID.String()))
	}

	if data.BuildFile, err = sdkbuildfile.Read(bytes.NewBuffer(buildData)); err != nil {
		return nil, fmt.Errorf("corrupted build file: %w", err)
	}

	return &data, nil
}
