package manifest

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type execContext struct {
	client   sdkservices.Services
	resolver resolver.Resolver

	projects     map[string]sdktypes.ProjectID
	integrations map[string]sdktypes.IntegrationID
	connections  map[string]sdktypes.ConnectionID
}

func (c *execContext) resolveProjectID(ctx context.Context, name string) (sdktypes.ProjectID, error) {
	if pid, ok := c.projects[name]; ok {
		return pid, nil
	}

	sdkName, err := sdktypes.ParseSymbol(name)
	if err != nil {
		return sdktypes.InvalidProjectID, err
	}

	p, err := c.client.Projects().GetByName(ctx, sdkName)
	if err != nil {
		return sdktypes.InvalidProjectID, err
	}

	pid := p.ID()
	c.projects[name] = pid
	return pid, nil
}

func (c *execContext) resolveIntegrationID(ctx context.Context, name string) (sdktypes.IntegrationID, error) {
	if iid, ok := c.integrations[name]; ok {
		return iid, nil
	}

	in, _, err := c.resolver.IntegrationNameOrID(ctx, name)
	if err != nil {
		return sdktypes.InvalidIntegrationID, err
	}

	iid := in.ID()
	c.integrations[name] = iid
	return iid, nil
}

func (c *execContext) resolveConnectionID(ctx context.Context, connID string) (sdktypes.ConnectionID, error) {
	if cid, ok := c.connections[connID]; ok {
		return cid, nil
	}

	conn, _, err := c.resolver.ConnectionNameOrID(ctx, connID, "")
	if err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	cid := conn.ID()
	c.connections[connID] = cid
	return cid, nil
}
