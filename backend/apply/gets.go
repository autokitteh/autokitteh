// Helper functions to resolve human parsable strings to concrete IDs.
package apply

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Getters struct{ Svcs sdkservices.Services }

func (g *Getters) GetConnectionID(ctx context.Context, project, x string) (sdktypes.ConnectionID, error) {
	if connID, err := sdktypes.StrictParseConnectionID(x); err == nil {
		// TODO: check if really exists.
		return connID, nil
	}

	if g.Svcs == nil {
		return nil, nil
	}

	var name string

	parts := strings.Split(x, "/")
	switch len(parts) {
	case 1:
		name = parts[0]
	case 2:
		project, name = parts[0], parts[1]
	default:
		return nil, fmt.Errorf(`invalid integration %q: must be "[<project>/]<name>"`, x)
	}

	pid, err := g.GetProjectID(ctx, project)
	if err != nil {
		return nil, err
	}

	if pid == nil {
		return nil, nil
	}

	cs, err := g.Svcs.Connections().List(ctx, sdkservices.ListConnectionsFilter{ProjectID: pid})
	if err != nil {
		return nil, fmt.Errorf("integrations.list: %w", err)
	}

	_, c := kittehs.FindFirst(cs, func(c sdktypes.Connection) bool {
		return sdktypes.GetConnectionName(c).String() == name
	})

	if c != nil {
		return sdktypes.GetConnectionID(c), nil
	}

	return nil, nil
}

func (g *Getters) GetIntegrationID(ctx context.Context, nameOrID string) (sdktypes.IntegrationID, error) {
	if iid, err := sdktypes.StrictParseIntegrationID(nameOrID); err == nil {
		// TODO: check if really exists.
		return iid, nil
	}

	is, err := g.Svcs.Integrations().List(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	for _, i := range is {
		if sdktypes.GetIntegrationUniqueName(i).String() == nameOrID {
			return sdktypes.GetIntegrationID(i), nil
		}
	}
	return nil, sdkerrors.ErrNotFound
}

func (g *Getters) GetProjectID(ctx context.Context, x string) (sdktypes.ProjectID, error) {
	if pid, err := sdktypes.StrictParseProjectID(x); err == nil {
		// TODO: check if exists.
		return pid, nil
	}

	if g.Svcs == nil {
		return nil, nil
	}

	n, err := sdktypes.StrictParseName(x)
	if err != nil {
		return nil, fmt.Errorf("name: %w", err)
	}

	p, err := g.Svcs.Projects().GetByName(ctx, n)
	if err != nil {
		return nil, fmt.Errorf("projects.get_by_name(%q): %w", n, err)
	}

	if p != nil {
		return sdktypes.GetProjectID(p), nil
	}

	return nil, nil
}
