package deployments

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type deployments struct {
	z  *zap.Logger
	db db.DB
}

func New(z *zap.Logger, db db.DB) sdkservices.Deployments {
	return &deployments{z: z, db: db}
}

func (d *deployments) Activate(ctx context.Context, id sdktypes.DeploymentID) error {
	return d.db.Transaction(ctx, func(tx db.DB) error {
		deployment, err := tx.GetDeployment(ctx, id)
		if err != nil {
			return fmt.Errorf("deployment: %w", err)
		}

		if deployment.State() == sdktypes.DeploymentStateActive {
			return nil
		}

		deployments, err := tx.ListDeployments(ctx, sdkservices.ListDeploymentsFilter{
			EnvID: deployment.EnvID(),
			State: sdktypes.DeploymentStateActive,
		})
		if err != nil {
			return fmt.Errorf("list active deployment: %w", err)
		}

		for _, d := range deployments {
			if err := drain(ctx, tx, d.ID()); err != nil {
				return err
			}
		}

		if err := tx.UpdateDeploymentState(ctx, id, sdktypes.DeploymentStateActive); err != nil {
			return fmt.Errorf("activate: %w", err)
		}

		return nil
	})
}

func (d *deployments) Test(ctx context.Context, id sdktypes.DeploymentID) error {
	return d.db.Transaction(ctx, func(tx db.DB) error {
		deployment, err := tx.GetDeployment(ctx, id)
		if err != nil {
			return fmt.Errorf("deployment: %w", err)
		}

		if deployment.State() == sdktypes.DeploymentStateTesting {
			return nil
		}

		if err := tx.UpdateDeploymentState(ctx, id, sdktypes.DeploymentStateTesting); err != nil {
			return fmt.Errorf("test: %w", err)
		}

		return nil
	})
}

func (d *deployments) Create(ctx context.Context, deployment sdktypes.Deployment) (sdktypes.DeploymentID, error) {
	deployment = deployment.WithNewID().WithState(sdktypes.DeploymentStateInactive)

	if err := d.db.CreateDeployment(ctx, deployment); err != nil {
		return sdktypes.InvalidDeploymentID, err
	}

	return deployment.ID(), nil
}

func deactivate(ctx context.Context, tx db.DB, id sdktypes.DeploymentID) error {
	_, nRunning, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
		DeploymentID: id,
		StateType:    sdktypes.SessionStateTypeRunning,
		CountOnly:    true,
	})
	if err != nil {
		return fmt.Errorf("sessions.count: %w", err)
	}

	_, nCreated, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
		DeploymentID: id,
		StateType:    sdktypes.SessionStateTypeCreated,
		CountOnly:    true,
	})
	if err != nil {
		return fmt.Errorf("sessions.count: %w", err)
	}

	if nRunning+nCreated > 0 {
		return fmt.Errorf("deployment still has pending sessions, drain first: %w", sdkerrors.ErrConflict)
	}

	return tx.UpdateDeploymentState(ctx, id, sdktypes.DeploymentStateInactive)
}

func (d *deployments) Deactivate(ctx context.Context, id sdktypes.DeploymentID) error {
	return d.db.Transaction(ctx, func(tx db.DB) error {
		return deactivate(ctx, tx, id)
	})
}

func drain(ctx context.Context, tx db.DB, id sdktypes.DeploymentID) error {
	if err := tx.UpdateDeploymentState(ctx, id, sdktypes.DeploymentStateDraining); err != nil {
		return err
	}

	if err := deactivate(ctx, tx, id); err != nil && !errors.Is(err, sdkerrors.ErrConflict) {
		return fmt.Errorf("deployment.deactivate(%v): %w", id, err)
	}

	return nil
}

func (d *deployments) Drain(ctx context.Context, id sdktypes.DeploymentID) error {
	return d.db.Transaction(ctx, func(tx db.DB) error {
		dep, err := tx.GetDeployment(ctx, id)
		if err != nil {
			return err
		}

		state := dep.State()
		if state != sdktypes.DeploymentStateActive && state != sdktypes.DeploymentStateTesting {
			return sdkerrors.ErrConflict
		}

		return drain(ctx, tx, id)
	})
}

func (d *deployments) Delete(ctx context.Context, id sdktypes.DeploymentID) error {
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
	return d.db.ListDeployments(ctx, filter)
}

func (d *deployments) Get(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	return sdkerrors.IgnoreNotFoundErr(d.db.GetDeployment(ctx, id))
}
