package dispatcher

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func resolveEnv(ctx context.Context, svcs *Services, env string) (envID sdktypes.EnvID, err error) {
	if env == "" {
		return sdktypes.InvalidEnvID, nil
	}

	parts := strings.SplitN(env, "/", 2)

	if len(parts) == 1 {
		if sdktypes.IsEnvID(parts[0]) {
			return sdktypes.ParseEnvID(parts[0])
		}

		var pid sdktypes.ProjectID

		if sdktypes.IsProjectID(parts[0]) {
			if pid, err = sdktypes.ParseProjectID(parts[0]); err != nil {
				return
			}
		} else {
			var name sdktypes.Symbol
			if name, err = sdktypes.ParseSymbol(parts[0]); err != nil {
				return
			}

			var p sdktypes.Project
			if p, err = svcs.Projects.GetByName(context.Background(), name); err != nil {
				return
			} else if !p.IsValid() {
				err = fmt.Errorf("project: %w", sdkerrors.ErrNotFound)
				return
			}

			pid = p.ID()
		}

		var envs []sdktypes.Env
		if envs, err = svcs.Envs.List(ctx, pid); err != nil {
			return
		}

		switch len(envs) {
		case 0:
			err = fmt.Errorf("env: %w", sdkerrors.ErrNotFound)
		case 1:
			envID = envs[0].ID()
		default:
			_, env := kittehs.FindFirst(envs, func(env sdktypes.Env) bool {
				return env.Name().String() == "default"
			})
			if env.IsValid() {
				envID = env.ID()
			} else {
				err = fmt.Errorf("env: %w", sdkerrors.ErrNotFound)
			}
		}
		return
	}

	var pid sdktypes.ProjectID

	if sdktypes.IsProjectID(parts[0]) {
		pid, err = sdktypes.ParseProjectID(parts[0])
	} else {
		var name sdktypes.Symbol
		if name, err = sdktypes.ParseSymbol(parts[0]); err != nil {
			return
		}

		var p sdktypes.Project
		if p, err = svcs.Projects.GetByName(context.Background(), name); p.IsValid() {
			pid = p.ID()
		}
	}

	if err != nil {
		return
	}

	if !pid.IsValid() {
		return sdktypes.InvalidEnvID, sdkerrors.ErrNotFound
	}

	var name sdktypes.Symbol
	if name, err = sdktypes.ParseSymbol(parts[1]); err != nil {
		return
	}

	var e sdktypes.Env
	if e, err = svcs.Envs.GetByName(ctx, pid, name); err != nil {
		return
	}

	if !e.IsValid() {
		err = sdkerrors.ErrNotFound
		return
	}

	envID = e.ID()
	return
}
