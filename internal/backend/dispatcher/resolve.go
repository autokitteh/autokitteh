package dispatcher

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Resolve an env string into an EnvID.
//   - If the env string is empty, it will return an InvalidEnvID.
//   - If the env string contains only a project name, it will return the default env for
//     that project. If the project does not have a default env, but have a single env,
//     it will return that env. If the project has more than a single env, it will fail.
//   - If the env string is of the form 'project/env', it will return that env.
func (d *Dispatcher) resolveEnv(ctx context.Context, env string) (sdktypes.EnvID, error) {
	if env == "" {
		return sdktypes.InvalidEnvID, nil
	}

	parts := strings.SplitN(env, "/", 2)

	if len(parts) == 1 {
		// Only a single part, meaning it's a project name.

		name, err := sdktypes.ParseSymbol(parts[0])
		if err != nil {
			return sdktypes.InvalidEnvID, err
		}

		p, err := d.svcs.Projects.GetByName(ctx, name)
		if err != nil {
			return sdktypes.InvalidEnvID, fmt.Errorf("project: %w", err)
		}

		envs, err := d.svcs.Envs.List(ctx, p.ID())
		if err != nil {
			return sdktypes.InvalidEnvID, err
		}

		switch len(envs) {
		case 0:
			// No env, nothing to return.
			return sdktypes.InvalidEnvID, fmt.Errorf("env: %w", sdkerrors.ErrNotFound)
		case 1:
			// Single env, return it.
			return envs[0].ID(), nil
		default:
			// More than one env, try to find the default one.
			_, env := kittehs.FindFirst(envs, func(env sdktypes.Env) bool {
				return env.Name().String() == "default"
			})

			if !env.IsValid() {
				return sdktypes.InvalidEnvID, fmt.Errorf("env: %w", sdkerrors.ErrNotFound)
			}

			return env.ID(), nil
		}
	}

	// Two parts, project and env.

	name, err := sdktypes.ParseSymbol(parts[0])
	if err != nil {
		return sdktypes.InvalidEnvID, err
	}

	p, err := d.svcs.Projects.GetByName(context.Background(), name)
	if err != nil {
		return sdktypes.InvalidEnvID, err
	}
	if !p.IsValid() {
		return sdktypes.InvalidEnvID, sdkerrors.ErrNotFound
	}

	if name, err = sdktypes.ParseSymbol(parts[1]); err != nil {
		return sdktypes.InvalidEnvID, err
	}

	e, err := d.svcs.Envs.GetByName(ctx, p.ID(), name)
	if err != nil {
		return sdktypes.InvalidEnvID, err
	}

	if !e.IsValid() {
		return sdktypes.InvalidEnvID, sdkerrors.ErrNotFound
	}

	return e.ID(), nil
}
