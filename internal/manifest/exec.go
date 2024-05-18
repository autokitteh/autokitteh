package manifest

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/manifest/internal/actions"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Execute(ctx context.Context, actions actions.Actions, client sdkservices.Services, log Log) (Effects, error) {
	execContext := execContext{
		client:   client,
		resolver: resolver.Resolver{Client: client},

		projects:     make(map[string]sdktypes.ProjectID),
		integrations: make(map[string]sdktypes.IntegrationID),
		envs:         make(map[string]sdktypes.EnvID),
		connections:  make(map[string]sdktypes.ConnectionID),
	}

	var effects []*Effect

	for _, action := range actions {
		effect, err := executeAction(ctx, action, &execContext)
		if err != nil {
			return nil, fmt.Errorf("action %s %s: %w", action.Type(), action.GetKey(), err)
		}

		effects = append(effects, effect)

		log.For(action.Type(), action).Printf("%v %s", effect.SubjectID.String(), effect.Type)
	}

	return effects, nil
}

func executeAction(ctx context.Context, action actions.Action, execContext *execContext) (*Effect, error) {
	switch action := action.(type) {
	case actions.CreateProjectAction:
		pid, err := execContext.client.Projects().Create(ctx, action.Project)
		if err != nil {
			return nil, err
		}

		execContext.projects[action.Project.Name().String()] = pid

		return &Effect{SubjectID: pid, Type: Created}, nil

	case actions.UpdateProjectAction:
		if err := execContext.client.Projects().Update(ctx, action.Project); err != nil {
			return nil, err
		}
		return &Effect{SubjectID: action.Project.ID(), Type: Created}, nil

	case actions.CreateConnectionAction:
		pid, err := execContext.resolveProjectID(ctx, action.ProjectKey)
		if err != nil {
			return nil, err
		}

		iid, err := execContext.resolveIntegrationID(action.IntegrationKey)
		if err != nil {
			return nil, err
		}

		conn := action.Connection.WithProjectID(pid).WithIntegrationID(iid)

		cid, err := execContext.client.Connections().Create(ctx, conn)
		if err != nil {
			return nil, err
		}

		execContext.connections[conn.Name().String()] = cid

		return &Effect{SubjectID: cid, Type: Created}, nil

	case actions.UpdateConnectionAction:
		err := execContext.client.Connections().Update(ctx, action.Connection)
		if err != nil {
			return nil, err
		}

		return &Effect{SubjectID: action.Connection.ID(), Type: Updated}, nil

	case actions.DeleteConnectionAction:
		err := execContext.client.Connections().Delete(ctx, action.ConnectionID)
		if err != nil {
			return nil, err
		}

		return &Effect{SubjectID: action.ConnectionID, Type: Deleted}, nil

	case actions.CreateEnvAction:
		pid, err := execContext.resolveProjectID(ctx, action.ProjectKey)
		if err != nil {
			return nil, err
		}

		env := action.Env.WithProjectID(pid)

		eid, err := execContext.client.Envs().Create(ctx, env)
		if err != nil {
			return nil, err
		}

		execContext.envs[env.Name().String()] = eid

		return &Effect{SubjectID: eid, Type: Created}, nil

	case actions.UpdateEnvAction:
		err := execContext.client.Envs().Update(ctx, action.Env)
		if err != nil {
			return nil, err
		}

		return &Effect{SubjectID: action.Env.ID(), Type: Updated}, nil

	case actions.DeleteEnvAction:
		if err := execContext.client.Envs().Remove(ctx, action.EnvID); err != nil {
			return nil, err
		}

		return &Effect{SubjectID: action.EnvID, Type: Deleted}, nil

	case actions.SetVarAction:
		var scopeID sdktypes.VarScopeID
		if action.Env != "" {
			eid, err := execContext.resolveEnvID(action.Env)
			if err != nil {
				return nil, err
			}

			scopeID = sdktypes.NewVarScopeID(eid)
		} else {
			cid, err := execContext.resolveConnectionID(action.Connection)
			if err != nil {
				return nil, err
			}

			scopeID = sdktypes.NewVarScopeID(cid)
		}

		v := action.Var.WithScopeID(scopeID)

		if err := execContext.client.Vars().Set(ctx, v); err != nil {
			return nil, err
		}

		return &Effect{SubjectID: scopeID, Type: Updated, Text: fmt.Sprintf("var %q updated", v.Name())}, nil

	case actions.DeleteVarAction:
		n, err := sdktypes.ParseSymbol(action.Name)
		if err != nil {
			return nil, err
		}
		if err = execContext.client.Vars().Delete(ctx, action.ScopeID, n); err != nil {
			return nil, err
		}

		return &Effect{SubjectID: action.ScopeID, Type: Updated, Text: fmt.Sprintf("var %q deleted", n)}, nil

	case actions.CreateTriggerAction:
		eid, err := execContext.resolveEnvID(action.EnvKey)
		if err != nil {
			return nil, err
		}

		cid, err := execContext.resolveConnectionID(action.ConnectionKey)
		if err != nil {
			return nil, err
		}

		t := action.Trigger.WithEnvID(eid).WithConnectionID(cid)

		tid, err := execContext.client.Triggers().Create(ctx, t)
		if err != nil {
			return nil, err
		}

		return &Effect{SubjectID: tid, Type: Created}, nil

	case actions.UpdateTriggerAction:
		if err := execContext.client.Triggers().Update(ctx, action.Trigger); err != nil {
			return nil, err
		}

		return &Effect{SubjectID: action.Trigger.ID(), Type: Updated}, nil

	case actions.DeleteTriggerAction:
		if err := execContext.client.Triggers().Delete(ctx, action.TriggerID); err != nil {
			return nil, err
		}

		return &Effect{SubjectID: action.TriggerID, Type: Deleted}, nil

	default:
		return nil, fmt.Errorf("unknown action type %T", action)
	}
}
