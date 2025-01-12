package deployments

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Draining will happen each interval + rand(jitter).
// Jitter is to prevent all instances from deactivating at the same time.
// Note that deactivation will happen if these are not set, as these are more
// to safeguard from a deployment being stuck in draining state.
type Config struct {
	AutoDrainingDeactivationInterval time.Duration `koanf:"draining_deactivation_interval"`
	AutoDrainingDeactivationJitter   time.Duration `koanf:"draining_deactivation_jitter"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		AutoDrainingDeactivationInterval: 5 * time.Minute,
		AutoDrainingDeactivationJitter:   1 * time.Minute,
	},
}

type deployments struct {
	l   *zap.Logger
	db  db.DB
	cfg *Config
}

func New(l *zap.Logger, cfg *Config, db db.DB, telemetry *telemetry.Telemetry) sdkservices.Deployments {
	initMetrics(telemetry)

	ds := &deployments{l: l, db: db, cfg: cfg}

	if cfg.AutoDrainingDeactivationInterval > 0 || cfg.AutoDrainingDeactivationJitter > 0 {
		go ds.Autodrain()
	}

	return ds
}

func (d *deployments) Autodrain() {
	l := d.l.With(
		zap.Duration("auto_drain_interval", d.cfg.AutoDrainingDeactivationInterval),
		zap.Duration("auto_drain_jitter", d.cfg.AutoDrainingDeactivationJitter),
	)
	l.Info("periodically auto-deactivating drained deployments")

	for {
		t := d.cfg.AutoDrainingDeactivationInterval
		t += time.Duration(rand.Float32() * float32(d.cfg.AutoDrainingDeactivationJitter))
		time.Sleep(t)

		n, err := d.db.DeactivateAllDrainedDeployments(context.Background())
		if err != nil {
			d.l.Error("auto-deactivation failed", zap.Error(err))
			continue
		}

		if n > 0 {
			d.l.Sugar().Infof("auto-deactivated %d drained deployments", n)
		}
	}
}

func (d *deployments) Activate(ctx context.Context, id sdktypes.DeploymentID) error {
	if err := authz.CheckContext(ctx, id, "write:activate"); err != nil {
		return err
	}

	l := d.l.With(zap.String("deployment_id", id.String()))
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
	if !filter.AnyIDSpecified() {
		filter.OrgID = authcontext.GetAuthnInferredOrgID(ctx)
	}

	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidDeploymentID, "read:list",
		authz.WithData("filter", filter),
		authz.WithAssociationWithID("project", filter.ProjectID),
		authz.WithAssociationWithID("build", filter.BuildID),
		authz.WithAssociationWithID("org", filter.OrgID),
	); err != nil {
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
