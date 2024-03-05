package manifest

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest/internal/actions"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Execute(ctx context.Context, actions actions.Actions, client sdkservices.Services, log Log) ([]sdktypes.ProjectID, error) {
	execContext := execContext{
		client:   client,
		log:      log,
		resolver: resolver.Resolver{Client: client},

		projects:     make(map[string]sdktypes.ProjectID),
		integrations: make(map[string]sdktypes.IntegrationID),
		envs:         make(map[string]sdktypes.EnvID),
		connections:  make(map[string]sdktypes.ConnectionID),
	}

	for _, action := range actions {
		if err := executeAction(ctx, action, &execContext); err != nil {
			return nil, fmt.Errorf("action %s %s: %w", action.Type(), action.GetKey(), err)
		}
	}

	return kittehs.MapValuesSortedByKeys(execContext.projects), nil
}

func executeAction(ctx context.Context, action actions.Action, execContext *execContext) error {
	log := execContext.log.For(action.Type(), action)

	switch action := action.(type) {
	case actions.CreateProjectAction:
		pid, err := execContext.client.Projects().Create(ctx, action.Project)
		if err != nil {
			return err
		}

		execContext.projects[action.Project.Name().String()] = pid

		log.Printf("created %q", pid)
	case actions.UpdateProjectAction:
		if err := execContext.client.Projects().Update(ctx, action.Project); err != nil {
			return err
		}

		log.Printf("updated")
	case actions.CreateConnectionAction:
		pid, err := execContext.resolveProjectID(ctx, action.ProjectKey)
		if err != nil {
			return err
		}

		iid, err := execContext.resolveIntegrationID(action.IntegrationKey)
		if err != nil {
			return err
		}

		conn := action.Connection.WithProjectID(pid).WithIntegrationID(iid)

		cid, err := execContext.client.Connections().Create(ctx, conn)
		if err != nil {
			return err
		}

		execContext.connections[conn.Name().String()] = cid

		log.Printf("created %q", cid)
	case actions.UpdateConnectionAction:
		err := execContext.client.Connections().Update(ctx, action.Connection)
		if err != nil {
			return err
		}

		log.Printf("updated")
	case actions.DeleteConnectionAction:
		err := execContext.client.Connections().Delete(ctx, action.ConnectionID)
		if err != nil {
			return err
		}

		log.Printf("updated")
	case actions.CreateEnvAction:
		pid, err := execContext.resolveProjectID(ctx, action.ProjectKey)
		if err != nil {
			return err
		}

		env := action.Env.WithProjectID(pid)

		eid, err := execContext.client.Envs().Create(ctx, env)
		if err != nil {
			return err
		}

		execContext.envs[env.Name().String()] = eid

		log.Printf("created %q", eid)
	case actions.UpdateEnvAction:
		err := execContext.client.Envs().Update(ctx, action.Env)
		if err != nil {
			return err
		}

		log.Printf("updated")
	case actions.DeleteEnvAction:
		if err := execContext.client.Envs().Remove(ctx, action.EnvID); err != nil {
			return err
		}

		log.Printf("deleted")
	case actions.SetEnvVarAction:
		eid, err := execContext.resolveEnvID(action.EnvKey)
		if err != nil {
			return err
		}

		v := action.EnvVar.WithEnvID(eid)

		if err = execContext.client.Envs().SetVar(ctx, v); err != nil {
			return err
		}

		log.Printf("set")
	case actions.DeleteEnvVarAction:
		n, err := sdktypes.ParseSymbol(action.Name)
		if err != nil {
			return err
		}
		if err = execContext.client.Envs().RemoveVar(ctx, action.EnvID, n); err != nil {
			return err
		}
		log.Printf("deleted")
	case actions.CreateTriggerAction:
		eid, err := execContext.resolveEnvID(action.EnvKey)
		if err != nil {
			return err
		}

		cid, err := execContext.resolveConnectionID(action.ConnectionKey)
		if err != nil {
			return err
		}

		t := action.Trigger.WithEnvID(eid).WithConnectionID(cid)

		tid, err := execContext.client.Triggers().Create(ctx, t)
		if err != nil {
			return err
		}

		log.Printf("created %q", tid)
	case actions.UpdateTriggerAction:
		if err := execContext.client.Triggers().Update(ctx, action.Trigger); err != nil {
			return err
		}

		log.Printf("updated")
	case actions.DeleteTriggerAction:
		if err := execContext.client.Triggers().Delete(ctx, action.TriggerID); err != nil {
			return err
		}
		log.Printf("deleted")
	default:
		return fmt.Errorf("unknown action type %T", action)
	}

	return nil
}
