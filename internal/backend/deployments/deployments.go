package deployments

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type deployments struct {
	z  *zap.Logger
	db db.DB
}

func New(z *zap.Logger, db db.DB, telemetry *telemetry.Telemetry) sdkservices.Deployments {
	initMetrics(telemetry)
	return &deployments{z: z, db: db}
}

func (d *deployments) Activate(ctx context.Context, id sdktypes.DeploymentID) error {
	if err := authz.CheckContext(ctx, id, "write:activate"); err != nil {
		return err
	}

	l := d.z.With(zap.String("deployment_id", id.String()))
	err := d.db.Transaction(ctx, func(tx db.DB) error {
		deployment, err := tx.GetDeployment(ctx, id)
		if err != nil {
			return fmt.Errorf("get deployment: %w", err)
		}

		if deployment.State() == sdktypes.DeploymentStateActive {
			return nil
		}

		deployments, err := tx.ListDeployments(ctx, sdkservices.ListDeploymentsFilter{
			ProjectID: deployment.ProjectID(),
			State:     sdktypes.DeploymentStateActive,
		})
		if err != nil {
			return fmt.Errorf("list active deployments: %w", err)
		}

		for _, d := range deployments {
			if d.State() != sdktypes.DeploymentStateInactive || d.State() != sdktypes.DeploymentStateDraining {
				if err := deactivate(ctx, tx, d.ID()); err != nil {
					return fmt.Errorf("deactivate deployment: %w", err)
				}
			}
		}

		if err := updateDeploymentState(ctx, tx, id, sdktypes.DeploymentStateActive); err != nil {
			return fmt.Errorf("activate deployment: %w", err)
		}

		return nil
	})
	if err != nil {
		l.Error("deployment activation failed", zap.Error(err))
		return err
	}

	l.Info("deployment activated")
	return nil
}

func (d *deployments) Test(ctx context.Context, id sdktypes.DeploymentID) error {
	if err := authz.CheckContext(ctx, id, "write:test"); err != nil {
		return err
	}

	return d.db.Transaction(ctx, func(tx db.DB) error {
		deployment, err := tx.GetDeployment(ctx, id)
		if err != nil {
			return fmt.Errorf("get deployment: %w", err)
		}

		if deployment.State() == sdktypes.DeploymentStateTesting {
			return nil
		}

		if _, err := tx.UpdateDeploymentState(ctx, id, sdktypes.DeploymentStateTesting); err != nil {
			return fmt.Errorf("test deployment: %w", err)
		}

		return nil
	})
}

func (d *deployments) Create(ctx context.Context, deployment sdktypes.Deployment) (sdktypes.DeploymentID, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidDeploymentID,
		"write:create",
		authz.WithData("deployment", deployment),
		authz.WithAssociationWithID("project", deployment.ProjectID()),
		authz.WithAssociationWithID("build", deployment.BuildID()),
	); err != nil {
		return sdktypes.InvalidDeploymentID, err
	}

	deployment = deployment.WithNewID().WithState(sdktypes.DeploymentStateInactive)

	if err := d.db.CreateDeployment(ctx, deployment); err != nil {
		return sdktypes.InvalidDeploymentID, err
	}

	deploymentsCreatedCounter.Add(ctx, 1)
	return deployment.ID(), nil
}

func (d *deployments) Deactivate(ctx context.Context, id sdktypes.DeploymentID) error {
	if err := authz.CheckContext(ctx, id, "write:deactivate"); err != nil {
		return err
	}

	return d.db.Transaction(ctx, func(tx db.DB) error { return deactivate(ctx, tx, id) })
}

func deactivate(ctx context.Context, tx db.DB, id sdktypes.DeploymentID) error {
	active, err := tx.DeploymentHasActiveSessions(ctx, id)
	if err != nil {
		return err
	}

	state := sdktypes.DeploymentStateInactive
	if active {
		state = sdktypes.DeploymentStateDraining
	}

	if err := updateDeploymentState(ctx, tx, id, state); err != nil {
		return err
	}

	return nil
}

func updateDeploymentState(ctx context.Context, db db.DB, id sdktypes.DeploymentID, state sdktypes.DeploymentState) error {
	oldState, err := db.UpdateDeploymentState(ctx, id, state)
	if err != nil {
		return err
	}

	updateStateCounter := func(state sdktypes.DeploymentState, val int64) {
		switch state {
		case sdktypes.DeploymentStateActive:
			deploymentsActiveGauge.Add(ctx, val)
		case sdktypes.DeploymentStateDraining:
			deploymentsDrainingGauge.Add(ctx, val)
		}
	}
	updateStateCounter(state, 1)
	updateStateCounter(oldState, -1)
	return nil
}

func (d *deployments) Delete(ctx context.Context, id sdktypes.DeploymentID) error {
	if err := authz.CheckContext(ctx, id, "delete:delete"); err != nil {
		return err
	}

	dep, err := d.db.GetDeployment(ctx, id)
	if err != nil {
		return err
	}

	if dep.State() != sdktypes.DeploymentStateInactive {
		return sdkerrors.ErrConflict
	}

	return d.db.DeleteDeployment(ctx, id)
}

func (d *deployments) List(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
	if err := authz.CheckContext(ctx, filter.ProjectID, "read:list-deployments", authz.WithData("filter", filter)); err != nil {
		return nil, err
	}

	return d.db.ListDeployments(ctx, filter)
}

func (d *deployments) Get(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	if err := authz.CheckContext(ctx, id, "read:get"); err != nil {
		return sdktypes.InvalidDeployment, err
	}

	return d.db.GetDeployment(ctx, id)
}
