package sessiondata

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuild/sdkbuildfile"
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
	Mappings    []sdktypes.Mapping      `json:"mappings"`
	Connections []sdktypes.Connection   `json:"connections"`
}

func retrieve[I sdktypes.ID, R any](ctx context.Context, z *zap.Logger, id I, f func(context.Context, I) (*R, error)) (*R, error) {
	r, err := f(ctx, id)

	if err != nil {
		z.DPanic("get", zap.Error(err), zap.String("id", id.String()))
		return nil, fmt.Errorf("get %q: %w", id, err)
	} else if r == nil {
		return nil, fmt.Errorf("%q not found", id)
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

func setDefaultMappings(ctx context.Context, data *Data) error {
	maps := kittehs.FilterNils(kittehs.Transform(data.Connections, func(desc sdktypes.Connection) sdktypes.Mapping {
		if _, m := kittehs.FindFirst(data.Mappings, func(m sdktypes.Mapping) bool {
			sameID := sdktypes.GetConnectionID(desc).String() == sdktypes.GetMappingConnectionID(m).String()
			sameName := sdktypes.GetConnectionName(desc).String() == sdktypes.GetMappingModuleName(m).String()
			return sameID || sameName
		}); m != nil {
			return nil
		}

		return kittehs.Must1(sdktypes.MappingFromProto(&sdktypes.MappingPB{
			MappingId:    sdktypes.NewMappingID().String(),
			EnvId:        sdktypes.GetEnvID(data.Env).String(),
			ConnectionId: sdktypes.GetConnectionID(desc).String(),
			ModuleName:   sdktypes.GetConnectionName(desc).String(),
		}))
	}))

	data.Mappings = append(data.Mappings, maps...)

	return nil
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

	if data.Deployment, err = retrieve(ctx, z, sdktypes.GetSessionDeploymentID(data.Session), svcs.Deployments.Get); err != nil {
		return nil, err
	}

	envID := sdktypes.GetDeploymentEnvID(data.Deployment)

	if data.Env, err = retrieve(ctx, z, envID, svcs.Envs.GetByID); err != nil {
		return nil, err
	}

	if data.ProjectID = sdktypes.GetEnvProjectID(data.Env); data.ProjectID == nil {
		return nil, fmt.Errorf("sessions can only run on projects")
	}

	if data.Connections, err = svcs.Connections.List(ctx, sdkservices.ListConnectionsFilter{ProjectID: data.ProjectID}); err != nil {
		return nil, fmt.Errorf("connections.list: %w", err)
	}

	if data.Mappings, err = svcs.Mappings.List(ctx, sdkservices.ListMappingsFilter{EnvID: envID}); err != nil {
		return nil, err
	}

	if err := setDefaultMappings(ctx, &data); err != nil {
		return nil, err
	}

	// TODO: merge mappings?

	if data.EnvVars, err = svcs.Envs.GetVars(ctx, nil, envID); err != nil {
		return nil, fmt.Errorf("get vars: %w", err)
	}

	buildID := sdktypes.GetDeploymentBuildID(data.Deployment)

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
