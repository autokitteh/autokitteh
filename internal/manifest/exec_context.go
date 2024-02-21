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
	log      Log
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

	sdkName, err := sdktypes.ParseName(name)
	if err != nil {
		return nil, err
	}

	p, err := c.client.Projects().GetByName(ctx, sdkName)
	if err != nil {
		return nil, err
	}

	pid := sdktypes.GetProjectID(p)
	c.projects[name] = pid
	return pid, nil
}

func (c *execContext) resolveIntegrationID(ctx context.Context, name string) (sdktypes.IntegrationID, error) {
	in, _, err := c.resolver.IntegrationNameOrID(name)
	iid := sdktypes.GetIntegrationID(in)
	c.integrations[name] = iid
	return iid, err
}

func (c *execContext) resolveEnvID(ctx context.Context, envID string) (sdktypes.EnvID, error) {
	if eid, ok := c.envs[envID]; ok {
		return eid, nil
	}

	proj, env, ok := strings.Cut(envID, "/")
	if !ok {
		return nil, fmt.Errorf("invalid env id %q", envID)
	}

	sdkEnv, _, err := c.resolver.EnvNameOrID(env, proj)
	if err != nil {
		return nil, err
	}

	eid := sdktypes.GetEnvID(sdkEnv)

	c.envs[envID] = eid

	return eid, nil
}

func (c *execContext) resolveConnectionID(ctx context.Context, connID string) (sdktypes.ConnectionID, error) {
	if cid, ok := c.connections[connID]; ok {
		return cid, nil
	}

	conn, _, err := c.resolver.ConnectionNameOrID(connID)
	if err != nil {
		return nil, err
	}

	cid := sdktypes.GetConnectionID(conn)

	c.connections[connID] = cid

	return cid, nil
}
