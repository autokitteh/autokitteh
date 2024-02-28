package manifest

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest/internal/actions"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const defaultEnvName = "default"

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

	name, err := sdktypes.ParseName(mproj.Name)
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

	if curr == nil {
		log.Printf("not found, will create")
		add(actions.CreateProjectAction{Key: mproj.GetKey(), Project: desired})
	} else {
		pid = sdktypes.GetProjectID(curr)

		log.Printf("found, id=%q", pid)

		desired = kittehs.Must1(desired.Update(func(p *sdktypes.ProjectPB) {
			p.ProjectId = pid.String()
		}))

		if sdktypes.Equal(curr, desired) {
			log.Printf("no changes needed")
		} else {
			log.Printf("not as desired, will update")

			add(actions.UpdateProjectAction{Key: mproj.GetKey(), Project: desired})
		}
	}

	// TODO: Remove all non-default environments.
	envActions, defaultEnvID, err := planDefaultEnv(ctx, mproj.Vars, client, mproj.Name, pid, optfns...)
	if err != nil {
		return nil, fmt.Errorf("envs: %w", err)
	}

	add(envActions...)

	connActions, conns, err := planConnections(ctx, mproj.Connections, client, mproj.Name, pid, optfns...)
	if err != nil {
		return nil, fmt.Errorf("connections: %w", err)
	}

	add(connActions...)

	triggerActions, err := planTriggers(ctx, mproj.Triggers, client, mproj.Name, pid, conns, defaultEnvID, optfns...)
	if err != nil {
		return nil, fmt.Errorf("triggers: %w", err)
	}

	add(triggerActions...)

	return acc, nil
}

func planDefaultEnv(ctx context.Context, mvars []*EnvVar, client sdkservices.Services, projName string, pid sdktypes.ProjectID, optfns ...Option) ([]actions.Action, sdktypes.EnvID, error) {
	envKeyer := stringKeyer(projName + "/" + defaultEnvName)
	opts := applyOptions(optfns)
	log := opts.log.For("env", envKeyer)

	if pid == nil && projName == "" {
		return nil, nil, errors.New("project must be set")
	}

	name := kittehs.Must1(sdktypes.ParseName(defaultEnvName))

	desired, err := sdktypes.EnvFromProto(&sdktypes.EnvPB{
		Name:      name.String(),
		ProjectId: pid.String(),
	})
	if err != nil {
		return nil, nil, err
	}

	var curr sdktypes.Env

	if pid != nil {
		if curr, err = client.Envs().GetByName(ctx, pid, name); err != nil {
			return nil, nil, fmt.Errorf("get env: %w", err)
		}
	}

	var (
		acc   []actions.Action
		add   = func(as ...actions.Action) { acc = append(acc, as...) }
		envID sdktypes.EnvID
	)

	if curr == nil {
		log.Printf("not found, will create")

		add(actions.CreateEnvAction{Key: envKeyer.GetKey(), ProjectKey: projName, Env: desired})
	} else {
		envID = sdktypes.GetEnvID(curr)

		log.Printf("found, id=%q", envID)

		desired = kittehs.Must1(desired.Update(func(p *sdktypes.EnvPB) {
			p.EnvId = envID.String()
		}))

		if sdktypes.Equal(curr, desired) {
			log.Printf("no changes needed")
		} else {
			log.Printf("not as desired, will update")

			add(actions.UpdateEnvAction{Key: envKeyer.GetKey(), Env: desired})
		}
	}

	var vars []sdktypes.EnvVar

	if envID != nil {
		if vars, err = client.Envs().GetVars(ctx, nil, envID); err != nil {
			return nil, nil, fmt.Errorf("get vars: %w", err)
		}
	}

	var mvarNames []string
	for _, mvar := range mvars {
		mvar := *mvar
		mvar.EnvKey = projName + "/" + defaultEnvName

		mvarNames = append(mvarNames, mvar.Name)

		_, v := kittehs.FindFirst(vars, func(v sdktypes.EnvVar) bool {
			return sdktypes.GetEnvVarName(v).String() == mvar.Name
		})

		val := mvar.Value
		if mvar.EnvVar != "" {
			if envVal, ok := os.LookupEnv(mvar.EnvVar); ok {
				val = envVal
			}
		}

		desired, err := sdktypes.EnvVarFromProto(&sdktypes.EnvVarPB{
			EnvId:    envID.String(),
			Name:     mvar.Name,
			Value:    val,
			IsSecret: mvar.IsSecret,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("invalid var: %w", err)
		}

		setAction := actions.SetEnvVarAction{Key: mvar.GetKey(), EnvKey: envKeyer.GetKey(), EnvVar: desired}

		log := opts.log.For("var", mvar)

		if v == nil {
			log("not found, will set")
			add(setAction)
		} else if envID == nil {
			// var was found, hence we must have an envID.
			sdklogger.Panic("envID is nil")
		} else {
			currVal := sdktypes.GetEnvVarValue(v)

			if sdktypes.IsEnvVarSecret(v) {
				if currVal, err = client.Envs().RevealVar(ctx, envID, sdktypes.GetEnvVarName(v)); err != nil {
					return nil, nil, fmt.Errorf("reveal var: %w", err)
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
		if name := sdktypes.GetEnvVarName(v).String(); !hasVar(name) {
			log.Printf("env var %q is not in the manifest, will delete", name)
			add(actions.DeleteEnvVarAction{Key: envKeyer.GetKey() + "/" + name, EnvID: envID, Name: name})
		}
	}

	return acc, envID, nil
}

func planConnections(ctx context.Context, mconns []*Connection, client sdkservices.Services, projName string, pid sdktypes.ProjectID, optfns ...Option) ([]actions.Action, []sdktypes.Connection, error) {
	opts := applyOptions(optfns)
	log := opts.log.For("project", stringKeyer(projName))

	var (
		acc       []actions.Action
		add       = func(as ...actions.Action) { acc = append(acc, as...) }
		conns     []sdktypes.Connection
		connNames []string
		err       error
	)

	if pid != nil && !opts.fromScratch {
		if conns, err = client.Connections().List(ctx, sdkservices.ListConnectionsFilter{ProjectID: pid}); err != nil {
			return nil, nil, fmt.Errorf("list connections: %w", err)
		}

		log.Printf("found %d connections", len(conns))
	}

	for _, mconn := range mconns {
		connNames = append(connNames, mconn.Name)

		if mconn.ProjectKey != "" {
			return nil, nil, errors.New("project must be empty")
		}

		mconn := *mconn
		mconn.ProjectKey = projName

		_, curr := kittehs.FindFirst(conns, func(c sdktypes.Connection) bool {
			return sdktypes.GetConnectionName(c).String() == mconn.Name
		})

		as, err := planConnection(ctx, &mconn, client, curr, optfns...)
		if err != nil {
			return nil, nil, fmt.Errorf("connection %q: %w", mconn.GetKey(), err)
		}

		add(as...)
	}

	if pid != nil {
		hasConn := kittehs.ContainedIn(connNames...)

		for _, conn := range conns {
			if name := sdktypes.GetConnectionName(conn); !hasConn(name.String()) {
				log.Printf("connection %q is not in the manifest, will delete", name)
				add(actions.DeleteConnectionAction{Key: projName + "/" + name.String(), ConnectionID: sdktypes.GetConnectionID(conn)})
			}
		}
	}

	return acc, conns, nil
}

func planConnection(ctx context.Context, mconn *Connection, client sdkservices.Services, curr sdktypes.Connection, optfns ...Option) ([]actions.Action, error) {
	opts := applyOptions(optfns)
	log := opts.log.For("connection", mconn)

	if curr == nil && mconn.ProjectKey == "" {
		return nil, errors.New("project must be set")
	}

	desired, err := sdktypes.ConnectionFromProto(&sdktypes.ConnectionPB{
		Name:             mconn.Name,
		IntegrationToken: mconn.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid: %w", err)
	}

	if curr == nil {
		log.Printf("not found, will create")
		return []actions.Action{actions.CreateConnectionAction{Key: mconn.GetKey(), ProjectKey: mconn.ProjectKey, IntegrationKey: mconn.IntegrationKey, Connection: desired}}, nil
	}

	desired = kittehs.Must1(desired.Update(func(pb *sdktypes.ConnectionPB) {
		pb.ConnectionId = sdktypes.GetConnectionID(curr).String()
		pb.IntegrationId = sdktypes.GetConnectionIntegrationID(curr).String()
		pb.ProjectId = sdktypes.GetConnectionProjectID(curr).String()
	}))

	if sdktypes.Equal(curr, desired) {
		log.Printf("no changes needed")
		return nil, nil
	}

	log.Printf("not as desired, will update")
	return []actions.Action{actions.UpdateConnectionAction{Key: mconn.GetKey(), Connection: desired}}, nil
}

func planTriggers(ctx context.Context, mtriggers []*Trigger, client sdkservices.Services, projName string, pid sdktypes.ProjectID, conns []sdktypes.Connection, defaultEnvID sdktypes.EnvID, optfns ...Option) ([]actions.Action, error) {
	opts := applyOptions(optfns)
	log := opts.log

	var (
		acc      []actions.Action
		add      = func(as ...actions.Action) { acc = append(acc, as...) }
		triggers []sdktypes.Trigger
	)

	if pid != nil && !opts.fromScratch {
		var err error
		if triggers, err = client.Triggers().List(ctx, sdkservices.ListTriggersFilter{ProjectID: pid}); err != nil {
			return nil, fmt.Errorf("list triggers: %w", err)
		}

		log.For("project", stringKeyer(projName)).Printf("found %d triggers", len(triggers))
	}

	connIDToName := kittehs.ListToMap(conns, func(c sdktypes.Connection) (string, string) {
		return sdktypes.GetConnectionID(c).String(), projName + "/" + sdktypes.GetConnectionName(c).String()
	})

	var matchedTriggerIDs []string

	for _, mtrigger := range mtriggers {
		mtrigger := *mtrigger
		mtrigger.EnvKey = projName + "/" + defaultEnvName
		mtrigger.ConnectionKey = projName + "/" + mtrigger.ConnectionKey

		log := log.For("trigger", mtrigger)

		_, curr := kittehs.FindFirst(triggers, func(t sdktypes.Trigger) bool {
			connName, ok := connIDToName[sdktypes.GetTriggerConnectionID(t).String()]
			if !ok {
				return false
			}

			if defaultEnvID == nil || sdktypes.GetTriggerEnvID(t).String() != defaultEnvID.String() {
				return false
			}

			return sdktypes.GetTriggerEventType(t) == mtrigger.EventType && connName == mtrigger.ConnectionKey
		})

		loc, err := sdktypes.ParseCodeLocation(mtrigger.Entrypoint)
		if err != nil {
			return nil, fmt.Errorf("trigger %q: invalid entrypoint: %w", mtrigger.GetKey(), err)
		}

		desired, err := sdktypes.TriggerFromProto(&sdktypes.TriggerPB{
			EventType:    mtrigger.EventType,
			CodeLocation: loc.ToProto(),
		})
		if err != nil {
			return nil, fmt.Errorf("trigger %q: invalid: %w", mtrigger.GetKey(), err)
		}

		if curr == nil {
			log.Printf("not found, will create")
			add(actions.CreateTriggerAction{Key: mtrigger.GetKey(), ConnectionKey: mtrigger.ConnectionKey, EnvKey: mtrigger.EnvKey, Trigger: desired})
		} else {
			matchedTriggerIDs = append(matchedTriggerIDs, sdktypes.GetTriggerID(curr).String())

			log.Printf("found, id=%q", sdktypes.GetTriggerID(curr))

			desired = kittehs.Must1(desired.Update(func(pb *sdktypes.TriggerPB) {
				pb.TriggerId = sdktypes.GetTriggerID(curr).String()
				pb.ConnectionId = sdktypes.GetTriggerConnectionID(curr).String()
				pb.EnvId = sdktypes.GetTriggerEnvID(curr).String()
			}))

			if sdktypes.Equal(curr, desired) {
				log.Printf("no changes needed")
			} else {
				log.Printf("not as desired, will update")
				add(actions.UpdateTriggerAction{Key: mtrigger.GetKey(), ConnectionKey: mtrigger.ConnectionKey, EnvKey: mtrigger.EnvKey, Trigger: desired})
			}
		}
	}

	hasTrigger := kittehs.ContainedIn(matchedTriggerIDs...)

	for _, trigger := range triggers {
		if tid := sdktypes.GetTriggerID(trigger); !hasTrigger(tid.String()) {
			log.Printf("trigger %q is not in the manifest, will delete", tid)
			add(actions.DeleteTriggerAction{Key: projName + "/" + tid.String(), TriggerID: tid})
		}
	}

	return acc, nil
}
