// Package resolver contains functions that resolve names and ID
// strings of autokitteh entities to their concrete SDK types.
package resolver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	separator = "/"
)

type Resolver struct {
	Client sdkservices.DBServices
}

// FIXME: move to sdkerrors
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

func translateError[O sdktypes.Object](err error, obj O, typ, idOrName string) error {
	// hack - rework
	what := "ID"
	if len(idOrName) > 4 && idOrName[4] != '_' {
		what = "name"
	}

	if err != nil {
		if !errors.Is(err, sdkerrors.ErrNotFound) {
			return fmt.Errorf("get %s %s %q: %w", typ, what, idOrName, err)
		}
		return err // not found
	}
	// no error. But most of the services.Get() methods filtering out notFound errors.
	// check sdktype.IsValid() to cover this case
	if !obj.IsValid() {
		return sdkerrors.ErrNotFound
	}
	return nil
}

// BuildID returns a build, based on the given ID.
// It does NOT accept empty input.
func (r Resolver) BuildID(ctx context.Context, id string) (b sdktypes.Build, bid sdktypes.BuildID, err error) {
	if id == "" {
		err = errors.New("missing build ID")
		return
	}

	if bid, err = sdktypes.StrictParseBuildID(id); err != nil {
		err = fmt.Errorf("invalid build ID %q: %w", id, err)
		return
	}

	b, err = r.Client.Builds().Get(ctx, bid)
	err = translateError(err, b, "build", id)
	return
}

// DeploymentID returns a deployment, based on the given ID.
// It does NOT accept empty input.
func (r Resolver) DeploymentID(ctx context.Context, id string) (d sdktypes.Deployment, did sdktypes.DeploymentID, err error) {
	if id == "" {
		err = errors.New("missing deployment ID")
		return
	}

	if did, err = sdktypes.Strict(sdktypes.ParseDeploymentID(id)); err != nil {
		err = fmt.Errorf("invalid deployment ID %q: %w", id, err)
		return
	}

	d, err = r.Client.Deployments().Get(ctx, did)
	err = translateError(err, d, "deployment", id)
	return
}

// ConnectionNameOrID returns a connection, based on the given name or
// ID. If the input is empty, we return nil but not an error.
func (r Resolver) ConnectionNameOrID(ctx context.Context, nameOrID, project string) (c sdktypes.Connection, cid sdktypes.ConnectionID, err error) {
	if nameOrID == "" {
		return
	}

	if sdktypes.IsConnectionID(nameOrID) {
		return r.connectionByID(ctx, nameOrID)
	}

	parts := strings.Split(nameOrID, separator)
	switch len(parts) {
	case 1:
		if project == "" {
			err = fmt.Errorf("invalid connection name %q: missing project prefix", nameOrID)
		} else {
			return r.connectionByFullName(ctx, project, parts[0], nameOrID)
		}
		return
	case 2:
		return r.connectionByFullName(ctx, parts[0], parts[1], nameOrID)
	default:
		err = fmt.Errorf("invalid connection name %q: too many parts", nameOrID)
		return
	}
}

func (r Resolver) connectionByID(ctx context.Context, id string) (c sdktypes.Connection, cid sdktypes.ConnectionID, err error) {
	if cid, err = sdktypes.StrictParseConnectionID(id); err != nil {
		err = fmt.Errorf("invalid connection ID %q: %w", id, err)
		return
	}

	if c, err = r.Client.Connections().Get(ctx, cid); err != nil {
		err = fmt.Errorf("get connection ID %q: %w", id, err)
		return
	}

	return
}

// TODO: add type and maybe id to sdkerrors.ErrNotFound and replace NotFoundError below
func (r Resolver) connectionByFullName(ctx context.Context, projNameOrID, connName, fullName string) (sdktypes.Connection, sdktypes.ConnectionID, error) {
	p, pid, err := r.ProjectNameOrID(ctx, projNameOrID)
	if err != nil {
		return sdktypes.InvalidConnection, sdktypes.InvalidConnectionID, err
	}
	if !p.IsValid() {
		return sdktypes.InvalidConnection, sdktypes.InvalidConnectionID, NotFoundError{Type: "project", Name: projNameOrID}
	}

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
func (r Resolver) EnvNameOrID(ctx context.Context, envNameOrID, projNameOrID string) (sdktypes.Env, sdktypes.EnvID, error) {
	if envNameOrID == "" {
		if projNameOrID == "" {
			return sdktypes.InvalidEnv, sdktypes.InvalidEnvID, nil
		} else {
			envNameOrID = "default"
		}
	}

	// Project.
	_, pid, err := r.ProjectNameOrID(ctx, projNameOrID)
	if err != nil {
		return sdktypes.InvalidEnv, sdktypes.InvalidEnvID, err
	}

	// Environment.
	if sdktypes.IsEnvID(envNameOrID) {
		return r.envByID(ctx, envNameOrID, projNameOrID, pid)
	}
	return r.envByName(ctx, envNameOrID, projNameOrID, pid)
}

func (r Resolver) envByID(ctx context.Context, envID, projNameOrID string, pid sdktypes.ProjectID) (e sdktypes.Env, eid sdktypes.EnvID, err error) {
	if eid, err = sdktypes.StrictParseEnvID(envID); err != nil {
		err = fmt.Errorf("invalid environment ID %q: %w", envID, err)
		return
	}

	e, err = r.Client.Envs().GetByID(ctx, eid)
	err = translateError(err, e, "environment", envID)
	if err != nil {
		return
	}

	if pid.IsValid() && pid != e.ProjectID() {
		err = fmt.Errorf("env ID %q doesn't belong to project %q", envID, projNameOrID)
		return
	}

	return
}

func (r Resolver) envByName(ctx context.Context, envName, projNameOrID string, pid sdktypes.ProjectID) (sdktypes.Env, sdktypes.EnvID, error) {
	parts := strings.Split(envName, separator)
	if len(parts) == 1 {
		return r.envByShortName(ctx, envName, pid)
	}
	return r.envByFullName(ctx, parts, projNameOrID, pid)
}

func (r Resolver) envByShortName(ctx context.Context, envName string, pid sdktypes.ProjectID) (e sdktypes.Env, eid sdktypes.EnvID, err error) {
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

	if e, err = r.Client.Envs().GetByName(ctx, pid, n); err != nil {
		err = fmt.Errorf("get environment name %q: %w", envName, err)
	}

	eid = e.ID()
	return
}

func (r Resolver) envByFullName(ctx context.Context, parts []string, projNameOrID string, pid sdktypes.ProjectID) (e sdktypes.Env, eid sdktypes.EnvID, err error) {
	prefix, envName := strings.Join(parts[:len(parts)-1], separator), parts[len(parts)-1]

	if e, eid, err = r.EnvNameOrID(ctx, envName, prefix); err != nil {
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
func (r Resolver) EventID(ctx context.Context, id string) (e sdktypes.Event, eid sdktypes.EventID, err error) {
	if id == "" {
		err = errors.New("missing event ID")
		return
	}

	if eid, err = sdktypes.Strict(sdktypes.ParseEventID(id)); err != nil {
		err = fmt.Errorf("invalid event ID %q: %w", id, err)
		return
	}

	e, err = r.Client.Events().Get(ctx, eid)
	err = translateError(err, e, "event", id)
	return
}

// IntegrationNameOrID returns an integration, based on the given
// name or ID. If the input is empty, we return nil but not an error.
func (r Resolver) IntegrationNameOrID(ctx context.Context, nameOrID string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	if nameOrID == "" {
		return sdktypes.InvalidIntegration, sdktypes.InvalidIntegrationID, nil
	}

	if sdktypes.IsIntegrationID(nameOrID) {
		return r.integrationByID(ctx, nameOrID)
	}

	return r.integrationByName(ctx, nameOrID)
}

func (r Resolver) integrationByID(ctx context.Context, id string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
	iid, err := sdktypes.StrictParseIntegrationID(id)
	if err != nil {
		return sdktypes.InvalidIntegration, sdktypes.InvalidIntegrationID, fmt.Errorf("invalid integration ID %q: %w", id, err)
	}

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

func (r Resolver) integrationByName(ctx context.Context, name string) (sdktypes.Integration, sdktypes.IntegrationID, error) {
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
func (r Resolver) TriggerID(ctx context.Context, id string) (t sdktypes.Trigger, tid sdktypes.TriggerID, err error) {
	if id == "" {
		err = errors.New("missing trigger ID")
		return
	}

	if tid, err = sdktypes.StrictParseTriggerID(id); err != nil {
		err = fmt.Errorf("invalid trigger ID %q: %w", id, err)
		return
	}

	t, err = r.Client.Triggers().Get(ctx, tid)
	err = translateError(err, t, "trigger", id)
	return
}

// ProjectNameOrID returns a project, based on the given name or ID.
// If the input is empty, we return nil but not an error.
func (r Resolver) ProjectNameOrID(ctx context.Context, nameOrID string) (sdktypes.Project, sdktypes.ProjectID, error) {
	if nameOrID == "" {
		return sdktypes.InvalidProject, sdktypes.InvalidProjectID, nil
	}

	if sdktypes.IsProjectID(nameOrID) {
		return r.projectByID(ctx, nameOrID)
	}

	return r.projectByName(ctx, nameOrID)
}

func (r Resolver) projectByID(ctx context.Context, id string) (p sdktypes.Project, pid sdktypes.ProjectID, err error) {
	if pid, err = sdktypes.StrictParseProjectID(id); err != nil {
		err = fmt.Errorf("invalid project ID %q: %w", id, err)
		return
	}
	p, err = r.Client.Projects().GetByID(ctx, pid)
	err = translateError(err, p, "project", id)

	return
}

func (r Resolver) projectByName(ctx context.Context, name string) (p sdktypes.Project, pid sdktypes.ProjectID, err error) {
	var n sdktypes.Symbol
	if n, err = sdktypes.Strict(sdktypes.ParseSymbol(name)); err != nil {
		err = fmt.Errorf("invalid project name %q: %w", name, err)
		return
	}

	p, err = r.Client.Projects().GetByName(ctx, n)
	err = translateError(err, p, "project", name)
	pid = p.ID()
	return
}

// SessionID returns a session, based on the given ID.
// If the input is empty, we return nil but not an error.
func (r Resolver) SessionID(ctx context.Context, id string) (s sdktypes.Session, sid sdktypes.SessionID, err error) {
	if id == "" {
		err = errors.New("missing session ID")
		return
	}

	if sid, err = sdktypes.StrictParseSessionID(id); err != nil {
		err = fmt.Errorf("invalid session ID %q: %w", id, err)
		return
	}

	s, err = r.Client.Sessions().Get(ctx, sid)
	err = translateError(err, s, "session", id)
	return
}
