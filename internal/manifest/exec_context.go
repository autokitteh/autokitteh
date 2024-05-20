package manifest

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type execContext struct {
	client   sdkservices.Services
	resolver resolver.Resolver

	projects     map[string]sdktypes.ProjectID
	integrations map[string]sdktypes.IntegrationID
	envs         map[string]sdktypes.EnvID
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

func (c *execContext) resolveIntegrationID(name string) (sdktypes.IntegrationID, error) {
	if iid, ok := c.integrations[name]; ok {
		return iid, nil
	}

	in, _, err := c.resolver.IntegrationNameOrID(name)
	iid := in.ID()
	c.integrations[name] = iid
	return iid, err
}

func (c *execContext) resolveEnvID(envID string) (sdktypes.EnvID, error) {
	if eid, ok := c.envs[envID]; ok {
		return eid, nil
	}

	proj, env, ok := strings.Cut(envID, "/")
	if !ok {
		return sdktypes.InvalidEnvID, fmt.Errorf("invalid env id %q", envID)
	}

	sdkEnv, _, err := c.resolver.EnvNameOrID(env, proj)
	if err != nil {
		return sdktypes.InvalidEnvID, err
	}

	eid := sdkEnv.ID()
	c.envs[envID] = eid
	return eid, nil
}

func (c *execContext) resolveConnectionID(connID string) (sdktypes.ConnectionID, error) {
	if cid, ok := c.connections[connID]; ok {
		return cid, nil
	}

	conn, _, err := c.resolver.ConnectionNameOrID(connID, "")
	if err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	cid := conn.ID()
	c.connections[connID] = cid
	return cid, nil
}
