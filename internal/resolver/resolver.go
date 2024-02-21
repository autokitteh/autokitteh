// Package resolver contains functions that resolve names and ID
// strings of autokitteh entities to their concrete SDK types.
package resolver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	separator = "/"
	timeout   = 10 * time.Second
)

type Resolver struct {
	Client sdkservices.Services
}

type NotFoundError struct {
	Type, Name string
}

var NotFoundErrorType *NotFoundError

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.Type, e.Name)
}

func limitedContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// BuildID returns a build, based on the given ID.
// It does NOT accept empty input.
//
// Subtle note: the build is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) BuildID(id string) (sdktypes.Build, sdktypes.BuildID, error) {
	if id == "" {
		return nil, nil, errors.New("missing build ID")
	}

	bid, err := sdktypes.StrictParseBuildID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid build ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	b, err := r.Client.Builds().Get(ctx, bid)
	if err != nil {
		return nil, nil, fmt.Errorf("get build ID %q: %w", id, err)
	}

	return b, bid, nil
}

// ConnectionNameOrID returns a connection, based on the given name or
// ID. If the input is empty, we return nil but not an error.
//
// Subtle note: the connection is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) ConnectionNameOrID(nameOrID string) (sdktypes.Connection, sdktypes.ConnectionID, error) {
	if nameOrID == "" {
		return nil, nil, nil
	}

	if sdktypes.IsID(nameOrID) {
		return r.connectionByID(nameOrID)
	}

	parts := strings.Split(nameOrID, separator)
	switch len(parts) {
	case 1:
		return nil, nil, fmt.Errorf("invalid connection name %q: missing project prefix", nameOrID)
	case 2:
		return r.connectionByFullName(parts[0], parts[1], nameOrID)
	default:
		return nil, nil, fmt.Errorf("invalid connection name %q: too many parts", nameOrID)
	}
}

func (r Resolver) connectionByID(id string) (sdktypes.Connection, sdktypes.ConnectionID, error) {
	cid, err := sdktypes.StrictParseConnectionID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid connection ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	c, err := r.Client.Connections().Get(ctx, cid)
	if err != nil {
		return nil, nil, fmt.Errorf("get connection ID %q: %w", id, err)
	}

	return c, cid, nil
}

func (r Resolver) connectionByFullName(projNameOrID, connName, fullName string) (sdktypes.Connection, sdktypes.ConnectionID, error) {
	p, pid, err := r.ProjectNameOrID(projNameOrID)
	if err != nil {
		return nil, nil, err
	}
	if p == nil {
		return nil, nil, NotFoundError{Type: "project", Name: projNameOrID}
	}

	ctx, cancel := limitedContext()
	defer cancel()

	f := sdkservices.ListConnectionsFilter{ProjectID: pid}
	cs, err := r.Client.Connections().List(ctx, f)
	if err != nil {
		return nil, nil, fmt.Errorf("list connections: %w", err)
	}

	for _, c := range cs {
		if sdktypes.GetConnectionName(c).String() == connName {
			return c, sdktypes.GetConnectionID(c), nil
		}
	}

	return nil, nil, NotFoundError{Type: "connection", Name: fullName}
}

// DeploymentID returns a deployment, based on the given ID.
// It does NOT accept empty input.
//
// Subtle note: the deployment is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) DeploymentID(id string) (sdktypes.Deployment, sdktypes.DeploymentID, error) {
	if id == "" {
		return nil, nil, errors.New("missing deployment ID")
	}

	did, err := sdktypes.StrictParseDeploymentID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid deployment ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	d, err := r.Client.Deployments().Get(ctx, did)
	if err != nil {
		return nil, nil, fmt.Errorf("get deployment ID %q: %w", id, err)
	}

	return d, did, nil
}

// EnvNameOrID returns an environment, based on the given environment
// and project names or IDs.
//
//   - If the input is empty, we return nil but not an error
//   - If the environment is specified as a name, the project is required
//   - If the environment is specified as a *full* name, the project is
//     optional, but if specified it must concur with the project prefix
//     in the environment name
//   - If the environment is specified as an ID, the project is optional,
//     but if specified it must concur with the environment's known project
//
// Subtle note: the environment is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) EnvNameOrID(envNameOrID, projNameOrID string) (sdktypes.Env, sdktypes.EnvID, error) {
	if envNameOrID == "" {
		return nil, nil, nil
	}

	// Project.
	_, pid, err := r.ProjectNameOrID(projNameOrID)
	if err != nil {
		return nil, nil, err
	}

	// Environment.
	if sdktypes.IsID(envNameOrID) {
		return r.envByID(envNameOrID, projNameOrID, pid)
	}
	return r.envByName(envNameOrID, projNameOrID, pid)
}

func (r Resolver) envByID(envID, projNameOrID string, pid sdktypes.ProjectID) (sdktypes.Env, sdktypes.EnvID, error) {
	eid, err := sdktypes.StrictParseEnvID(envID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid environment ID %q: %w", envID, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	e, err := r.Client.Envs().GetByID(ctx, eid)
	if err != nil {
		return nil, eid, fmt.Errorf("get environment ID %q: %w", envID, err)
	}

	if pid != nil && pid.String() != sdktypes.GetEnvProjectID(e).String() {
		return nil, eid, fmt.Errorf("env ID %q doesn't belong to project %q", envID, projNameOrID)
	}

	return e, eid, nil
}

func (r Resolver) envByName(envName, projNameOrID string, pid sdktypes.ProjectID) (sdktypes.Env, sdktypes.EnvID, error) {
	parts := strings.Split(envName, separator)
	if len(parts) == 1 {
		return r.envByShortName(envName, projNameOrID, pid)
	}
	return r.envByFullName(parts, projNameOrID, pid)
}

func (r Resolver) envByShortName(envName, projNameOrID string, pid sdktypes.ProjectID) (sdktypes.Env, sdktypes.EnvID, error) {
	if pid == nil {
		return nil, nil, fmt.Errorf("invalid environment name %q: missing project prefix", envName)
	}

	n, err := sdktypes.StrictParseName(envName)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid environment name %q: %w", envName, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	e, err := r.Client.Envs().GetByName(ctx, pid, n)
	if err != nil {
		return nil, nil, fmt.Errorf("get environment name %q: %w", envName, err)
	}

	return e, sdktypes.GetEnvID(e), nil
}

func (r Resolver) envByFullName(parts []string, projNameOrID string, pid sdktypes.ProjectID) (sdktypes.Env, sdktypes.EnvID, error) {
	prefix, envName := strings.Join(parts[:len(parts)-1], separator), parts[len(parts)-1]

	e, eid, err := r.EnvNameOrID(envName, prefix)
	if err != nil {
		return nil, nil, err
	}

	// Sanity check: the original project must match the prefix.
	if pid != nil && pid.String() != sdktypes.GetEnvProjectID(e).String() {
		return nil, eid, fmt.Errorf("env %q doesn't belong to project %q", prefix, projNameOrID)
	}

	return e, eid, nil
}

// EventID returns an event, based on the given ID.
// It does NOT accept empty input.
//
// Subtle note: the event is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) EventID(id string) (sdktypes.Event, sdktypes.EventID, error) {
	if id == "" {
		return nil, nil, errors.New("missing event ID")
	}

	eid, err := sdktypes.StrictParseEventID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid event ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	e, err := r.Client.Events().Get(ctx, eid)
	if err != nil {
		return nil, nil, fmt.Errorf("get event ID %q: %w", id, err)
	}

	return e, eid, nil
}

// IntegrationNameOrID returns an integration, based on the given
// name or ID. If the input is empty, we return nil but not an error.
//
// Subtle note: the integration is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) IntegrationNameOrID(nameOrID string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	if nameOrID == "" {
		return nil, nil, nil
	}

	if sdktypes.IsID(nameOrID) {
		return r.integrationByID(nameOrID)
	}

	return r.integrationByName(nameOrID)
}

func (r Resolver) integrationByID(id string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	iid, err := sdktypes.StrictParseIntegrationID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid integration ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	is, err := r.Client.Integrations().List(ctx, "")
	if err != nil {
		return nil, nil, fmt.Errorf("list integrations: %w", err)
	}

	for _, i := range is {
		if sdktypes.GetIntegrationID(i).String() == iid.String() {
			return i, iid, nil
		}
	}

	return nil, iid, nil
}

func (r Resolver) integrationByName(name string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	ctx, cancel := limitedContext()
	defer cancel()

	is, err := r.Client.Integrations().List(ctx, name)
	if err != nil {
		return nil, nil, fmt.Errorf("list integrations: %w", err)
	}

	for _, i := range is {
		if sdktypes.GetIntegrationUniqueName(i).String() == name {
			return i, sdktypes.GetIntegrationID(i), nil
		}
		if sdktypes.GetIntegrationDisplayName(i) == name {
			return i, sdktypes.GetIntegrationID(i), nil
		}
	}

	return nil, nil, nil
}

// TriggerID returns a trigger, based on the given ID.
// If the input is empty, we return nil but not an error.
//
// Subtle note: the trigger is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) TriggerID(id string) (sdktypes.Trigger, sdktypes.TriggerID, error) {
	if id == "" {
		return nil, nil, fmt.Errorf("missing trigger ID")
	}

	mid, err := sdktypes.StrictParseTriggerID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid trigger ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	m, err := r.Client.Triggers().Get(ctx, mid)
	if err != nil {
		return nil, nil, fmt.Errorf("get trigger ID %q: %w", id, err)
	}

	return m, mid, nil
}

// ProjectNameOrID returns a project, based on the given name or ID.
// If the input is empty, we return nil but not an error.
//
// Subtle note: the project is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) ProjectNameOrID(nameOrID string) (sdktypes.Project, sdktypes.ProjectID, error) {
	if nameOrID == "" {
		return nil, nil, nil
	}

	if sdktypes.IsID(nameOrID) {
		return r.projectByID(nameOrID)
	}

	return r.projectByName(nameOrID)
}

func (r Resolver) projectByID(id string) (sdktypes.Project, sdktypes.ProjectID, error) {
	pid, err := sdktypes.StrictParseProjectID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid project ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	p, err := r.Client.Projects().GetByID(ctx, pid)
	if err != nil {
		return nil, pid, fmt.Errorf("get project ID %q: %w", id, err)
	}

	return p, pid, nil
}

func (r Resolver) projectByName(name string) (sdktypes.Project, sdktypes.ProjectID, error) {
	n, err := sdktypes.StrictParseName(name)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid project name %q: %w", name, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	p, err := r.Client.Projects().GetByName(ctx, n)
	if err != nil {
		return nil, nil, fmt.Errorf("get project name %q: %w", name, err)
	}

	return p, sdktypes.GetProjectID(p), nil
}

// SessionID returns a session, based on the given ID.
// If the input is empty, we return nil but not an error.
//
// Subtle note: the session is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) SessionID(id string) (sdktypes.Session, sdktypes.SessionID, error) {
	if id == "" {
		return nil, nil, errors.New("missing session ID")
	}

	sid, err := sdktypes.StrictParseSessionID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid session ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	s, err := r.Client.Sessions().Get(ctx, sid)
	if err != nil {
		return nil, nil, fmt.Errorf("get session ID %q: %w", id, err)
	}

	return s, sid, nil
}
