package deployments

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
	l := d.z.With(zap.String("deployment_id", id.String()))
	err := d.db.Transaction(ctx, func(tx db.DB) error {
		deployment, err := tx.GetDeployment(ctx, id)
		if err != nil {
			return kittehs.ErrorWithPrefix("get deployment", err)
		}

		if deployment.State() == sdktypes.DeploymentStateActive {
			return nil
		}

		deployments, err := tx.ListDeployments(ctx, sdkservices.ListDeploymentsFilter{
			EnvID: deployment.EnvID(),
			State: sdktypes.DeploymentStateActive,
		})
		if err != nil {
			return kittehs.ErrorWithPrefix("list active deployments", err)
		}

		for _, d := range deployments {
			if err := drain(ctx, tx, d.ID()); err != nil {
				return kittehs.ErrorWithPrefix("drain deployment", err)
			}
		}

		if err := updateDeploymentState(ctx, tx, id, sdktypes.DeploymentStateActive); err != nil {
			return kittehs.ErrorWithPrefix("activate deployment", err)
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
	return d.db.Transaction(ctx, func(tx db.DB) error {
		deployment, err := tx.GetDeployment(ctx, id)
		if err != nil {
			return kittehs.ErrorWithPrefix("get deployment", err)
		}

		if deployment.State() == sdktypes.DeploymentStateTesting {
			return nil
		}

		if _, err := tx.UpdateDeploymentState(ctx, id, sdktypes.DeploymentStateTesting); err != nil {
			return kittehs.ErrorWithPrefix("test deployment", err)
		}

		return nil
	})
}

func (d *deployments) Create(ctx context.Context, deployment sdktypes.Deployment) (sdktypes.DeploymentID, error) {
	deployment = deployment.WithNewID().WithState(sdktypes.DeploymentStateInactive)

	if err := d.db.CreateDeployment(ctx, deployment); err != nil {
		return sdktypes.InvalidDeploymentID, err
	}

	deploymentsCreatedCounter.Add(ctx, 1)
	return deployment.ID(), nil
}

func deactivate(ctx context.Context, tx db.DB, id sdktypes.DeploymentID) error {
	// TODO: single query?
	resultRunning, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
		DeploymentID: id,
		StateType:    sdktypes.SessionStateTypeRunning,
		CountOnly:    true,
	})
	if err != nil {
		return kittehs.ErrorWithPrefix("count running sessions", err)
	}

	resultCreated, err := tx.ListSessions(ctx, sdkservices.ListSessionsFilter{
		DeploymentID: id,
		StateType:    sdktypes.SessionStateTypeCreated,
		CountOnly:    true,
	})
	if err != nil {
		return kittehs.ErrorWithPrefix("count created sessions", err)
	}

	if resultRunning.TotalCount+resultCreated.TotalCount > 0 {
		return kittehs.ErrorWithPrefix("deployment still has pending sessions, drain first", sdkerrors.ErrConflict)
	}

	return updateDeploymentState(ctx, tx, id, sdktypes.DeploymentStateInactive)
}

func (d *deployments) Deactivate(ctx context.Context, id sdktypes.DeploymentID) error {
	return d.db.Transaction(ctx, func(tx db.DB) error {
		return deactivate(ctx, tx, id)
	})
}

func drain(ctx context.Context, tx db.DB, id sdktypes.DeploymentID) error {
	if err := updateDeploymentState(ctx, tx, id, sdktypes.DeploymentStateDraining); err != nil {
		return err
	}

	if err := deactivate(ctx, tx, id); err != nil && !errors.Is(err, sdkerrors.ErrConflict) {
		return fmt.Errorf("deployment.deactivate(%v): %w", id, err)
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
	return d.db.GetDeployment(ctx, id)
}
