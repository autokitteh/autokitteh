package manifest

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest/internal/actions"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
		curr, err = sdkerrors.IgnoreNotFoundErr(client.Projects().GetByName(ctx, name))
		if err != nil {
			return nil, fmt.Errorf("get project: %w", err)
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
	pvarsActions, err := planProjectVars(ctx, mproj.Vars, client, mproj.Name, pid, optfns...)
	if err != nil {
		return nil, fmt.Errorf("vars: %w", err)
	}

	add(pvarsActions...)

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

func planProjectVars(ctx context.Context, mvars []*Var, client sdkservices.Services, projName string, pid sdktypes.ProjectID, optfns ...Option) ([]actions.Action, error) {
	opts := applyOptions(optfns)

	sid := sdktypes.NewVarScopeID(pid)

	var (
		acc       []actions.Action
		add       = func(as ...actions.Action) { acc = append(acc, as...) }
		mvarNames []string
		err       error
	)

	var vars sdktypes.Vars
	if sid.IsValid() {
		if vars, err = client.Vars().Get(ctx, sid); err != nil {
			return nil, fmt.Errorf("get vars: %w", err)
		}
	}

	for _, mvar := range mvars {
		mvar := *mvar
		mvar.ParentKey = projName

		mvarNames = append(mvarNames, mvar.Name)

		v := vars.GetByString(mvar.Name)

		n, err := sdktypes.StrictParseSymbol(mvar.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid var name: %w", err)
		}

		desired := sdktypes.NewVar(n).SetValue(mvar.Value).SetSecret(mvar.Secret).WithScopeID(sid)

		setAction := actions.SetVarAction{Key: mvar.GetKey(), Project: projName, Var: desired}

		log := opts.log.For("var", mvar)

		if !v.IsValid() {
			log("not found, will set")
			add(setAction)
		} else {
			currVal := v.Value()

			if currVal != mvar.Value {
				log("differs, will set")
				add(setAction)
			}
		}
	}

	hasVar := kittehs.ContainedIn(mvarNames...)
	for _, v := range vars {
		if name := v.Name().String(); !hasVar(name) {
			log.Printf("env var %q is not in the manifest, will delete", name)
			add(actions.DeleteVarAction{Key: projName + "/" + name, ScopeID: sid, Name: name})
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

		cid, as, err := planConnection(&mconn, curr, optfns...)
		if err != nil {
			return nil, fmt.Errorf("connection %q: %w", mconn.GetKey(), err)
		}

		add(as...)

		sid := sdktypes.NewVarScopeID(cid)

		var cvars []sdktypes.Var
		if cid.IsValid() {
			if cvars, err = client.Vars().Get(ctx, sid); err != nil {
				return nil, fmt.Errorf("get connection vars: %w", err)
			}
		}

		if as, err = planConnectionVars(mconn, cid, cvars, optfns...); err != nil {
			return nil, fmt.Errorf("connection var %q: %w", mconn.GetKey(), err)
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

func planConnectionVars(mconn Connection, cid sdktypes.ConnectionID, cvars sdktypes.Vars, optfns ...Option) (acts []actions.Action, err error) {
	opts := applyOptions(optfns)

	var handled []sdktypes.Symbol

	for _, mvar := range mconn.Vars {
		mvar.ParentKey = mconn.GetKey()

		log := opts.log.For("var", mvar)

		n, err := sdktypes.ParseSymbol(mvar.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid var name: %w", err)
		}

		want := sdktypes.NewVar(n).SetValue(mvar.Value).SetSecret(mvar.Secret).WithScopeID(sdktypes.NewVarScopeID(cid))

		got := cvars.Get(want.Name())

		handled = append(handled, n)

		if got.Equal(want) {
			log.Printf("no change needed")
			continue
		}

		if got.IsValid() {
			log.Printf("changed, will update")
		} else {
			log.Printf("not found, will set")
		}

		acts = append(acts, actions.SetVarAction{Key: mvar.GetKey(), Connection: mconn.GetKey(), Var: want})
	}

	// Remove connection vars not in the manifest.
	if opts.rmUnusedConnVars {
		hasVar := kittehs.ContainedIn(handled...)
		removed := kittehs.Filter(cvars, func(cvar sdktypes.Var) bool {
			return !hasVar(cvar.Name())
		})

		acts = append(acts, kittehs.Transform(removed, func(cvar sdktypes.Var) actions.Action {
			log.Printf("connection var %q not in the manifest, will delete", cvar.Name())
			return actions.DeleteVarAction{Key: mconn.GetKey() + "/" + cvar.Name().String(), ScopeID: sdktypes.NewVarScopeID(cid), Name: cvar.Name().String()}
		})...)
	}

	return acts, nil
}

func planConnection(mconn *Connection, curr sdktypes.Connection, optfns ...Option) (sdktypes.ConnectionID, []actions.Action, error) {
	opts := applyOptions(optfns)
	log := opts.log.For("connection", mconn)

	if !curr.IsValid() && mconn.ProjectKey == "" {
		return sdktypes.InvalidConnectionID, nil, errors.New("project must be set")
	}

	desired, err := sdktypes.ConnectionFromProto(&sdktypes.ConnectionPB{
		Name: mconn.Name,
	})
	if err != nil {
		return sdktypes.InvalidConnectionID, nil, fmt.Errorf("invalid connection: %w", err)
	}

	if !curr.IsValid() {
		log.Printf("not found, will create")
		actions := []actions.Action{
			actions.CreateConnectionAction{
				Key:            mconn.GetKey(),
				ProjectKey:     mconn.ProjectKey,
				IntegrationKey: mconn.IntegrationKey,
				Connection:     desired,
			},
		}
		return sdktypes.InvalidConnectionID, actions, nil
	}

	desired = desired.
		WithID(curr.ID()).
		WithIntegrationID(curr.IntegrationID()).
		WithProjectID(curr.ProjectID())

	if curr.WithoutGeneratedFields().Equal(desired) {
		log.Printf("no changes needed")
		return curr.ID(), nil, nil
	}

	log.Printf("not as desired, will update")
	actions := []actions.Action{
		actions.UpdateConnectionAction{
			Key:        mconn.GetKey(),
			Connection: desired,
		},
	}
	return curr.ID(), actions, nil
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
		triggers, err = client.Triggers().List(ctx, sdkservices.ListTriggersFilter{ProjectID: pid})
		if err != nil {
			return nil, fmt.Errorf("list triggers: %w", err)
		}

		log.For("project", stringKeyer(projName)).Printf("found %d triggers", len(triggers))
	}

	var matchedTriggerIDs []string

	for _, mtrigger := range mtriggers {
		mtrigger := *mtrigger
		mtrigger.ProjectKey = projName

		if mtrigger.ConnectionKey != nil {
			*mtrigger.ConnectionKey = projName + "/" + *mtrigger.ConnectionKey
		}

		log := log.For("trigger", mtrigger)

		_, curr := kittehs.FindFirst(triggers, func(t sdktypes.Trigger) bool {
			return t.Name().String() == mtrigger.Name
		})

		loc, err := sdktypes.ParseCodeLocation(mtrigger.Call)
		if err != nil {
			return nil, fmt.Errorf("trigger %q: invalid entrypoint: %w", mtrigger.GetKey(), err)
		}

		desired, err := sdktypes.TriggerFromProto(&sdktypes.TriggerPB{
			Filter:       mtrigger.Filter,
			EventType:    mtrigger.EventType,
			CodeLocation: loc.ToProto(),
			Name:         mtrigger.Name,
			ProjectId:    pid.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("trigger %q: invalid: %w", mtrigger.GetKey(), err)
		}

		if mtrigger.Webhook != nil || mtrigger.Type == "webhook" {
			if mtrigger.Type != "" && mtrigger.Type != "webhook" {
				return nil, fmt.Errorf("trigger %q: type %q is not supported for webhook", mtrigger.GetKey(), mtrigger.Type)
			}

			desired = desired.WithWebhook()
		}

		if mtrigger.ConnectionKey != nil || mtrigger.Type == "connection" {
			if mtrigger.Type != "" && mtrigger.Type != "connection" {
				return nil, fmt.Errorf("trigger %q: type %q is not supported for connection", mtrigger.GetKey(), mtrigger.Type)
			}

			desired = desired.WithSourceType(sdktypes.TriggerSourceTypeConnection)
		}

		if mtrigger.Schedule != nil || mtrigger.Type == "schedule" {
			if mtrigger.Type != "" && mtrigger.Type != "schedule" {
				return nil, fmt.Errorf("trigger %q: type %q is not supported for schedule", mtrigger.GetKey(), mtrigger.Type)
			}

			desired = desired.WithSchedule(*mtrigger.Schedule)
		}

		if desired.SourceType() == sdktypes.TriggerSourceTypeUnspecified {
			return nil, fmt.Errorf("trigger %q: concrete type not specified", mtrigger.GetKey())
		}

		if !curr.IsValid() {
			log.Printf("not found, will create")
			add(actions.CreateTriggerAction{Key: mtrigger.GetKey(), ConnectionKey: mtrigger.ConnectionKey, ProjectKey: mtrigger.ProjectKey, Trigger: desired})
		} else {
			matchedTriggerIDs = append(matchedTriggerIDs, curr.ID().String())
			log.Printf("found, id=%q", curr.ID())

			if desired.SourceType() != curr.SourceType() {
				return nil, fmt.Errorf("trigger %q: source type cannot be updated", mtrigger.GetKey())
			}

			curr = curr.WithWebhookSlug("")

			// TODO: `curr' may have actual connectionID (if not a scheduler trigger)
			// while for `desired' we need to resolve it from connectionKey
			// therefore for now `cur' and `desired' will have the same connectionID for comparison
			desired = desired.
				WithID(curr.ID()).
				WithName(curr.Name()).
				WithProjectID(curr.ProjectID())

			if curr.Equal(desired) {
				log.Printf("no changes needed")
			} else {
				log.Printf("not as desired, will update")
				add(actions.UpdateTriggerAction{Key: mtrigger.GetKey(), ConnectionKey: mtrigger.ConnectionKey, ProjectKey: mtrigger.ProjectKey, Trigger: desired})
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
