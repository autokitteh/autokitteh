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

var NotFoundErrorType = new(NotFoundError)

func (e NotFoundError) Error() string {
	name := e.Name
	if name != "" {
		name = fmt.Sprintf(" %q", name)
	}
	return e.Type + name + " not found"
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
func (r Resolver) BuildID(id string) (b sdktypes.Build, bid sdktypes.BuildID, err error) {
	if id == "" {
		err = errors.New("missing build ID")
		return
	}

	if bid, err = sdktypes.StrictParseBuildID(id); err != nil {
		err = fmt.Errorf("invalid build ID %q: %w", id, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if b, err = r.Client.Builds().Get(ctx, bid); err != nil {
		err = fmt.Errorf("get build ID %q: %w", id, err)
		return
	}

	return
}

// ConnectionNameOrID returns a connection, based on the given name or
// ID. If the input is empty, we return nil but not an error.
//
// Subtle note: the connection is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) ConnectionNameOrID(nameOrID, project string) (c sdktypes.Connection, cid sdktypes.ConnectionID, err error) {
	if nameOrID == "" {
		return
	}

	if sdktypes.IsConnectionID(nameOrID) {
		return r.connectionByID(nameOrID)
	}

	parts := strings.Split(nameOrID, separator)
	switch len(parts) {
	case 1:
		if project == "" {
			err = fmt.Errorf("invalid connection name %q: missing project prefix", nameOrID)
		} else {
			return r.connectionByFullName(project, parts[0], nameOrID)
		}
		return
	case 2:
		return r.connectionByFullName(parts[0], parts[1], nameOrID)
	default:
		err = fmt.Errorf("invalid connection name %q: too many parts", nameOrID)
		return
	}
}

func (r Resolver) connectionByID(id string) (c sdktypes.Connection, cid sdktypes.ConnectionID, err error) {
	if cid, err = sdktypes.StrictParseConnectionID(id); err != nil {
		err = fmt.Errorf("invalid connection ID %q: %w", id, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if c, err = r.Client.Connections().Get(ctx, cid); err != nil {
		err = fmt.Errorf("get connection ID %q: %w", id, err)
		return
	}

	return
}

func (r Resolver) connectionByFullName(projNameOrID, connName, fullName string) (sdktypes.Connection, sdktypes.ConnectionID, error) {
	p, pid, err := r.ProjectNameOrID(projNameOrID)
	if err != nil {
		return sdktypes.InvalidConnection, sdktypes.InvalidConnectionID, err
	}
	if !p.IsValid() {
		return sdktypes.InvalidConnection, sdktypes.InvalidConnectionID, NotFoundError{Type: "project", Name: projNameOrID}
	}

	ctx, cancel := limitedContext()
	defer cancel()

	f := sdkservices.ListConnectionsFilter{ProjectID: pid}
	cs, err := r.Client.Connections().List(ctx, f)
	if err != nil {
		return sdktypes.InvalidConnection, sdktypes.InvalidConnectionID, fmt.Errorf("list connections: %w", err)
	}

	for _, c := range cs {
		if c.Name().String() == connName {
			return c, c.ID(), nil
		}
	}

	return sdktypes.InvalidConnection, sdktypes.InvalidConnectionID, NotFoundError{Type: "connection", Name: fullName}
}

// DeploymentID returns a deployment, based on the given ID.
// It does NOT accept empty input.
//
// Subtle note: the deployment is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) DeploymentID(id string) (d sdktypes.Deployment, did sdktypes.DeploymentID, err error) {
	if id == "" {
		err = errors.New("missing deployment ID")
		return
	}

	if did, err = sdktypes.Strict(sdktypes.ParseDeploymentID(id)); err != nil {
		err = fmt.Errorf("invalid deployment ID %q: %w", id, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if d, err = r.Client.Deployments().Get(ctx, did); err != nil {
		err = fmt.Errorf("get deployment ID %q: %w", id, err)
		return
	}

	return
}

// EnvNameOrID returns an environment, based on the given environment
// and project names or IDs.
//
//   - If the input is empty, we return nil but not an error
//   - If the environment is empty but the project isn't, we try to
//     resolve the environment as the default one for the project
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
		if projNameOrID == "" {
			return sdktypes.InvalidEnv, sdktypes.InvalidEnvID, nil
		} else {
			envNameOrID = "default"
		}
	}

	// Project.
	_, pid, err := r.ProjectNameOrID(projNameOrID)
	if err != nil {
		return sdktypes.InvalidEnv, sdktypes.InvalidEnvID, err
	}

	// Environment.
	if sdktypes.IsEnvID(envNameOrID) {
		return r.envByID(envNameOrID, projNameOrID, pid)
	}
	return r.envByName(envNameOrID, projNameOrID, pid)
}

func (r Resolver) envByID(envID, projNameOrID string, pid sdktypes.ProjectID) (e sdktypes.Env, eid sdktypes.EnvID, err error) {
	if eid, err = sdktypes.StrictParseEnvID(envID); err != nil {
		err = fmt.Errorf("invalid environment ID %q: %w", envID, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if e, err = r.Client.Envs().GetByID(ctx, eid); err != nil {
		err = fmt.Errorf("get environment ID %q: %w", envID, err)
		return
	}

	if pid.IsValid() && pid != e.ProjectID() {
		err = fmt.Errorf("env ID %q doesn't belong to project %q", envID, projNameOrID)
		return
	}

	return
}

func (r Resolver) envByName(envName, projNameOrID string, pid sdktypes.ProjectID) (sdktypes.Env, sdktypes.EnvID, error) {
	parts := strings.Split(envName, separator)
	if len(parts) == 1 {
		return r.envByShortName(envName, pid)
	}
	return r.envByFullName(parts, projNameOrID, pid)
}

func (r Resolver) envByShortName(envName string, pid sdktypes.ProjectID) (e sdktypes.Env, eid sdktypes.EnvID, err error) {
	if !pid.IsValid() {
		err = fmt.Errorf("invalid environment name %q: missing project prefix", envName)
		return
	}

	var n sdktypes.Symbol
	n, err = sdktypes.Strict(sdktypes.ParseSymbol(envName))
	if err != nil {
		err = fmt.Errorf("invalid environment name %q: %w", envName, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if e, err = r.Client.Envs().GetByName(ctx, pid, n); err != nil {
		err = fmt.Errorf("get environment name %q: %w", envName, err)
	}

	eid = e.ID()
	return
}

func (r Resolver) envByFullName(parts []string, projNameOrID string, pid sdktypes.ProjectID) (e sdktypes.Env, eid sdktypes.EnvID, err error) {
	prefix, envName := strings.Join(parts[:len(parts)-1], separator), parts[len(parts)-1]

	if e, eid, err = r.EnvNameOrID(envName, prefix); err != nil {
		return
	}

	// Sanity check: the original project must match the prefix.
	if pid.IsValid() && pid != e.ProjectID() {
		err = fmt.Errorf("env %q doesn't belong to project %q", envName, projNameOrID)
		return
	}

	return
}

// EventID returns an event, based on the given ID.
// It does NOT accept empty input.
//
// Subtle note: the event is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) EventID(id string) (e sdktypes.Event, eid sdktypes.EventID, err error) {
	if id == "" {
		err = errors.New("missing event ID")
		return
	}

	if eid, err = sdktypes.Strict(sdktypes.ParseEventID(id)); err != nil {
		err = fmt.Errorf("invalid event ID %q: %w", id, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if e, err = r.Client.Events().Get(ctx, eid); err != nil {
		err = fmt.Errorf("get event ID %q: %w", id, err)
		return
	}

	return
}

// IntegrationNameOrID returns an integration, based on the given
// name or ID. If the input is empty, we return nil but not an error.
//
// Subtle note: the integration is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) IntegrationNameOrID(nameOrID string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	if nameOrID == "" {
		return sdktypes.InvalidIntegration, sdktypes.InvalidIntegrationID, nil
	}

	if sdktypes.IsIntegrationID(nameOrID) {
		return r.integrationByID(nameOrID)
	}

	return r.integrationByName(nameOrID)
}

func (r Resolver) integrationByID(id string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	iid, err := sdktypes.StrictParseIntegrationID(id)
	if err != nil {
		return sdktypes.InvalidIntegration, sdktypes.InvalidIntegrationID, fmt.Errorf("invalid integration ID %q: %w", id, err)
	}

	ctx, cancel := limitedContext()
	defer cancel()

	is, err := r.Client.Integrations().List(ctx, "")
	if err != nil {
		return sdktypes.InvalidIntegration, sdktypes.InvalidIntegrationID, fmt.Errorf("list integrations: %w", err)
	}

	for _, i := range is {
		if i.ID() == iid {
			return i, iid, nil
		}
	}

	return sdktypes.InvalidIntegration, iid, nil
}

func (r Resolver) integrationByName(name string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	ctx, cancel := limitedContext()
	defer cancel()

	is, err := r.Client.Integrations().List(ctx, name)
	if err != nil {
		return sdktypes.InvalidIntegration, sdktypes.InvalidIntegrationID, fmt.Errorf("list integrations: %w", err)
	}

	for _, i := range is {
		if i.UniqueName().String() == name {
			return i, i.ID(), nil
		}
		if i.DisplayName() == name {
			return i, i.ID(), nil
		}
	}

	return sdktypes.InvalidIntegration, sdktypes.InvalidIntegrationID, nil
}

// TriggerID returns a trigger, based on the given ID.
// If the input is empty, we return nil but not an error.
//
// Subtle note: the trigger is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) TriggerID(id string) (t sdktypes.Trigger, tid sdktypes.TriggerID, err error) {
	if id == "" {
		err = errors.New("missing trigger ID")
		return
	}

	if tid, err = sdktypes.StrictParseTriggerID(id); err != nil {
		err = fmt.Errorf("invalid trigger ID %q: %w", id, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if t, err = r.Client.Triggers().Get(ctx, tid); err != nil {
		err = fmt.Errorf("get trigger ID %q: %w", id, err)
		return
	}

	return
}

// ProjectNameOrID returns a project, based on the given name or ID.
// If the input is empty, we return nil but not an error.
//
// Subtle note: the project is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) ProjectNameOrID(nameOrID string) (sdktypes.Project, sdktypes.ProjectID, error) {
	if nameOrID == "" {
		return sdktypes.InvalidProject, sdktypes.InvalidProjectID, nil
	}

	if sdktypes.IsProjectID(nameOrID) {
		return r.projectByID(nameOrID)
	}

	return r.projectByName(nameOrID)
}

func (r Resolver) projectByID(id string) (p sdktypes.Project, pid sdktypes.ProjectID, err error) {
	if pid, err = sdktypes.StrictParseProjectID(id); err != nil {
		err = fmt.Errorf("invalid project ID %q: %w", id, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	p, err = r.Client.Projects().GetByID(ctx, pid)
	if err != nil {
		err = fmt.Errorf("get project ID %q: %w", id, err)
		return
	}

	return
}

func (r Resolver) projectByName(name string) (p sdktypes.Project, pid sdktypes.ProjectID, err error) {
	var n sdktypes.Symbol
	if n, err = sdktypes.Strict(sdktypes.ParseSymbol(name)); err != nil {
		err = fmt.Errorf("invalid project name %q: %w", name, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if p, err = r.Client.Projects().GetByName(ctx, n); err != nil {
		err = fmt.Errorf("get project name %q: %w", name, err)
	}

	pid = p.ID()
	return
}

// SessionID returns a session, based on the given ID.
// If the input is empty, we return nil but not an error.
//
// Subtle note: the session is guaranteed to exist only if the FIRST
// return value is non-nil. Example: if the input is a valid ID,
// but it doesn't actually exist, we return (nil, ID, nil).
func (r Resolver) SessionID(id string) (s sdktypes.Session, sid sdktypes.SessionID, err error) {
	if id == "" {
		err = errors.New("missing session ID")
		return
	}

	if sid, err = sdktypes.StrictParseSessionID(id); err != nil {
		err = fmt.Errorf("invalid session ID %q: %w", id, err)
		return
	}

	ctx, cancel := limitedContext()
	defer cancel()

	if s, err = r.Client.Sessions().Get(ctx, sid); err != nil {
		err = fmt.Errorf("get session ID %q: %w", id, err)
	}

	return
}
