package manifest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest/internal/actions"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	defaultEnvName = "default"
)

var ErrUnsupportedManifestVersion = errors.New("unsupported manifest version")

func Plan(
	ctx context.Context,
	manifest *Manifest,
	client sdkservices.Services,
	optfns ...Option,
) (actions.Actions, error) {
	if manifest.Version != Version {
		return nil, fmt.Errorf("%w: got %v, expected %v", ErrUnsupportedManifestVersion, manifest.Version, Version)
	}

	var actions []actions.Action

	if mproj := manifest.Project; mproj != nil {
		projectActions, err := planProject(ctx, mproj, client, optfns...)
		if err != nil {
			return nil, fmt.Errorf("project %q: %w", mproj.GetKey(), err)
		}

		actions = append(projectActions, actions...)
	}

	return actions, nil
}

func planProject(ctx context.Context, mproj *Project, client sdkservices.Services, optfns ...Option) ([]actions.Action, error) {
	opts := applyOptions(optfns)
	log := opts.log.For("project", mproj)

	if opts.projectName != "" {
		mproj.Name = opts.projectName
	}

	if mproj.Name == "" {
		return nil, errors.New("project name must be specified")
	}

	name, err := sdktypes.ParseSymbol(mproj.Name)
	if err != nil {
		return nil, err
	}

	var curr sdktypes.Project
	if !opts.fromScratch {
		if curr, err = client.Projects().GetByName(ctx, name); err != nil {
			return nil, fmt.Errorf("get: %w", err)
		}
	}

	var (
		acc []actions.Action
		add = func(as ...actions.Action) { acc = append(acc, as...) }
		pid sdktypes.ProjectID
	)

	desired, err := sdktypes.ProjectFromProto(&sdktypes.ProjectPB{
		Name: mproj.Name,
	})
	if err != nil {
		return nil, err
	}

	if !curr.IsValid() {
		log.Printf("not found, will create")
		add(actions.CreateProjectAction{Key: mproj.GetKey(), Project: desired})
	} else {
		pid = curr.ID()

		log.Printf("found, id=%q", pid)

		desired = desired.WithID(pid)

		if curr.Equal(desired) {
			log.Printf("no changes needed")
		} else {
			log.Printf("not as desired, will update")

			add(actions.UpdateProjectAction{Key: mproj.GetKey(), Project: desired})
		}
	}

	// TODO: Remove all non-default environments.
	envActions, err := planDefaultEnv(ctx, mproj.Vars, client, mproj.Name, pid, optfns...)
	if err != nil {
		return nil, fmt.Errorf("envs: %w", err)
	}

	add(envActions...)

	connActions, err := planConnections(ctx, mproj.Connections, client, mproj.Name, pid, optfns...)
	if err != nil {
		return nil, fmt.Errorf("connections: %w", err)
	}

	add(connActions...)

	triggerActions, err := planTriggers(ctx, mproj.Triggers, client, mproj.Name, pid, optfns...)
	if err != nil {
		return nil, fmt.Errorf("triggers: %w", err)
	}

	add(triggerActions...)

	return acc, nil
}

func planDefaultEnv(ctx context.Context, mvars []*EnvVar, client sdkservices.Services, projName string, pid sdktypes.ProjectID, optfns ...Option) ([]actions.Action, error) {
	envKeyer := stringKeyer(projName + "/" + defaultEnvName)
	opts := applyOptions(optfns)
	log := opts.log.For("env", envKeyer)

	if !pid.IsValid() && projName == "" {
		return nil, errors.New("project must be set")
	}

	name := kittehs.Must1(sdktypes.ParseSymbol(defaultEnvName))

	desired, err := sdktypes.EnvFromProto(&sdktypes.EnvPB{
		Name:      name.String(),
		ProjectId: pid.String(),
	})
	if err != nil {
		return nil, err
	}

	var curr sdktypes.Env

	if pid.IsValid() {
		if curr, err = client.Envs().GetByName(ctx, pid, name); err != nil {
			return nil, fmt.Errorf("get env: %w", err)
		}
	}

	var (
		acc   []actions.Action
		add   = func(as ...actions.Action) { acc = append(acc, as...) }
		envID sdktypes.EnvID
	)

	if curr.IsValid() {
		envID = curr.ID()

		log.Printf("found, id=%q", envID)

		desired = desired.WithID(envID)

		if curr.Equal(desired) {
			log.Printf("no changes needed")
		} else {
			log.Printf("not as desired, will update")

			add(actions.UpdateEnvAction{Key: envKeyer.GetKey(), Env: desired})
		}
	}

	var vars []sdktypes.EnvVar

	if envID.IsValid() {
		if vars, err = client.Envs().GetVars(ctx, nil, envID); err != nil {
			return nil, fmt.Errorf("get vars: %w", err)
		}
	}

	var mvarNames []string
	for _, mvar := range mvars {
		mvar := *mvar
		mvar.EnvKey = projName + "/" + defaultEnvName

		mvarNames = append(mvarNames, mvar.Name)

		_, v := kittehs.FindFirst(vars, func(v sdktypes.EnvVar) bool {
			return v.Symbol().String() == mvar.Name
		})

		val := mvar.Value
		envVar := mvar.EnvVar

		if mvar.IsSecret {
			if val != "" {
				return nil, errors.New("value cannot be specified for secrets")
			}

			if envVar == "" {
				envVar = strings.ToUpper(mvar.Name)
			}
		}

		if envVar != "" {
			if envVal, ok := os.LookupEnv(envVar); ok {
				val = envVal
			} else if mvar.IsSecret {
				return nil, fmt.Errorf("env var %q is secret and not found in the environment", envVar)
			}
		}

		desired, err := sdktypes.EnvVarFromProto(&sdktypes.EnvVarPB{
			EnvId:    envID.String(),
			Name:     mvar.Name,
			Value:    val,
			IsSecret: mvar.IsSecret,
		})
		if err != nil {
			return nil, fmt.Errorf("invalid var: %w", err)
		}

		setAction := actions.SetEnvVarAction{Key: mvar.GetKey(), EnvKey: envKeyer.GetKey(), EnvVar: desired}

		log := opts.log.For("var", mvar)

		if !v.IsValid() {
			log("not found, will set")
			add(setAction)
		} else if !envID.IsValid() {
			// var was found, hence we must have an envID.
			sdklogger.Panic("envID is nil")
		} else {
			currVal := v.Value()

			if v.IsSecret() {
				if currVal, err = client.Envs().RevealVar(ctx, envID, v.Symbol()); err != nil {
					return nil, fmt.Errorf("reveal var: %w", err)
				}
			}

			if currVal != mvar.Value {
				log("differs, will set")
				add(setAction)
			}
		}
	}

	hasVar := kittehs.ContainedIn(mvarNames...)
	for _, v := range vars {
		if name := v.Symbol().String(); !hasVar(name) {
			log.Printf("env var %q is not in the manifest, will delete", name)
			add(actions.DeleteEnvVarAction{Key: envKeyer.GetKey() + "/" + name, EnvID: envID, Name: name})
		}
	}

	return acc, nil
}

func planConnections(ctx context.Context, mconns []*Connection, client sdkservices.Services, projName string, pid sdktypes.ProjectID, optfns ...Option) ([]actions.Action, error) {
	opts := applyOptions(optfns)
	log := opts.log.For("project", stringKeyer(projName))

	var (
		acc       []actions.Action
		add       = func(as ...actions.Action) { acc = append(acc, as...) }
		conns     []sdktypes.Connection
		connNames []string
		err       error
	)

	if pid.IsValid() && !opts.fromScratch {
		if conns, err = client.Connections().List(ctx, sdkservices.ListConnectionsFilter{ProjectID: pid}); err != nil {
			return nil, fmt.Errorf("list connections: %w", err)
		}

		log.Printf("found %d connections", len(conns))
	}

	for _, mconn := range mconns {
		connNames = append(connNames, mconn.Name)

		if mconn.ProjectKey != "" {
			return nil, errors.New("project must be empty")
		}

		mconn := *mconn
		mconn.ProjectKey = projName

		_, curr := kittehs.FindFirst(conns, func(c sdktypes.Connection) bool {
			return c.Name().String() == mconn.Name
		})

		as, err := planConnection(&mconn, curr, optfns...)
		if err != nil {
			return nil, fmt.Errorf("connection %q: %w", mconn.GetKey(), err)
		}

		add(as...)
	}

	if pid.IsValid() {
		hasConn := kittehs.ContainedIn(connNames...)

		for _, conn := range conns {
			if name := conn.Name(); !hasConn(name.String()) {
				log.Printf("connection %q is not in the manifest, will delete", name)
				add(actions.DeleteConnectionAction{Key: projName + "/" + name.String(), ConnectionID: conn.ID()})
			}
		}
	}

	return acc, nil
}

func planConnection(mconn *Connection, curr sdktypes.Connection, optfns ...Option) ([]actions.Action, error) {
	opts := applyOptions(optfns)
	log := opts.log.For("connection", mconn)

	if !curr.IsValid() && mconn.ProjectKey == "" {
		return nil, errors.New("project must be set")
	}

	desired, err := sdktypes.ConnectionFromProto(&sdktypes.ConnectionPB{
		Name:             mconn.Name,
		IntegrationToken: mconn.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid: %w", err)
	}

	if !curr.IsValid() {
		log.Printf("not found, will create")
		return []actions.Action{actions.CreateConnectionAction{Key: mconn.GetKey(), ProjectKey: mconn.ProjectKey, IntegrationKey: mconn.IntegrationKey, Connection: desired}}, nil
	}

	desired = desired.
		WithID(curr.ID()).
		WithIntegrationID(curr.IntegrationID()).
		WithProjectID(curr.ProjectID())

	if curr.Equal(desired) {
		log.Printf("no changes needed")
		return nil, nil
	}

	log.Printf("not as desired, will update")
	return []actions.Action{actions.UpdateConnectionAction{Key: mconn.GetKey(), Connection: desired}}, nil
}

func planTriggers(ctx context.Context, mtriggers []*Trigger, client sdkservices.Services, projName string, pid sdktypes.ProjectID, optfns ...Option) ([]actions.Action, error) {
	opts := applyOptions(optfns)
	log := opts.log

	var (
		acc      []actions.Action
		add      = func(as ...actions.Action) { acc = append(acc, as...) }
		triggers []sdktypes.Trigger
	)

	if pid.IsValid() && !opts.fromScratch {
		var err error
		if triggers, err = client.Triggers().List(ctx, sdkservices.ListTriggersFilter{ProjectID: pid}); err != nil {
			return nil, fmt.Errorf("list triggers: %w", err)
		}

		log.For("project", stringKeyer(projName)).Printf("found %d triggers", len(triggers))
	}

	var matchedTriggerIDs []string

	for _, mtrigger := range mtriggers {
		mtrigger := *mtrigger
		mtrigger.EnvKey = projName + "/" + defaultEnvName
		mtrigger.ConnectionKey = projName + "/" + mtrigger.ConnectionKey

		log := log.For("trigger", mtrigger)

		_, curr := kittehs.FindFirst(triggers, func(t sdktypes.Trigger) bool {
			return mtrigger.Name != "" && t.Name() == mtrigger.Name
		})

		ep := mtrigger.Entrypoint
		if ep == "" {
			ep = mtrigger.Call
		}
		loc, err := sdktypes.ParseCodeLocation(ep)
		if err != nil {
			return nil, fmt.Errorf("trigger %q: invalid entrypoint: %w", mtrigger.GetKey(), err)
		}

		data, err := kittehs.TransformMapValuesError(mtrigger.Data, sdktypes.WrapValue)
		if err != nil {
			return nil, fmt.Errorf("trigger %q: invalid additional data: %w", mtrigger.GetKey(), err)
		}

		desired, err := sdktypes.TriggerFromProto(&sdktypes.TriggerPB{
			Filter:       mtrigger.Filter,
			EventType:    mtrigger.EventType,
			CodeLocation: loc.ToProto(),
			Data:         kittehs.TransformMapValues(data, sdktypes.ToProto),
			Name:         mtrigger.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("trigger %q: invalid: %w", mtrigger.GetKey(), err)
		}

		if !curr.IsValid() {
			log.Printf("not found, will create")
			add(actions.CreateTriggerAction{Key: mtrigger.GetKey(), ConnectionKey: mtrigger.ConnectionKey, EnvKey: mtrigger.EnvKey, Trigger: desired})
		} else {
			matchedTriggerIDs = append(matchedTriggerIDs, curr.ID().String())

			log.Printf("found, id=%q", curr.ID())

			desired = desired.
				WithID(curr.ID()).
				WithName(curr.Name()).
				WithConnectionID(curr.ConnectionID()).
				WithEnvID(curr.EnvID())

			if curr.Equal(desired) {
				log.Printf("no changes needed")
			} else {
				log.Printf("not as desired, will update")
				add(actions.UpdateTriggerAction{Key: mtrigger.GetKey(), ConnectionKey: mtrigger.ConnectionKey, EnvKey: mtrigger.EnvKey, Trigger: desired})
			}
		}
	}

	hasTrigger := kittehs.ContainedIn(matchedTriggerIDs...)

	for _, trigger := range triggers {
		if tid := trigger.ID(); !hasTrigger(tid.String()) {
			log.Printf("trigger %q is not in the manifest, will delete", tid)
			add(actions.DeleteTriggerAction{Key: projName + "/" + tid.String(), TriggerID: tid})
		}
	}

	return acc, nil
}
