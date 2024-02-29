package dbgorm

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func triggerToRecord(ctx context.Context, tx *tx, trigger sdktypes.Trigger) (*scheme.Trigger, error) {
	connID := sdktypes.GetTriggerConnectionID(trigger)

	conn, err := tx.GetConnection(ctx, connID)
	if err != nil {
		return nil, fmt.Errorf("get trigger connection: %w", err)
	}

	projID := sdktypes.GetConnectionProjectID(conn)

	envID := sdktypes.GetTriggerEnvID(trigger)
	if envID != nil {
		env, err := tx.GetEnvByID(ctx, envID)
		if err != nil {
			return nil, fmt.Errorf("get trigger env: %w", err)
		}

		if projID.String() != sdktypes.GetEnvProjectID(env).String() {
			return nil, fmt.Errorf("env and connection project mismatch: %s != %s", projID.String(), sdktypes.GetEnvProjectID(env).String())
		}
	}

	return &scheme.Trigger{
		TriggerID:    sdktypes.GetTriggerID(trigger).String(),
		EnvID:        envID.String(),
		ProjectID:    projID.String(),
		ConnectionID: connID.String(),
		EventType:    sdktypes.GetTriggerEventType(trigger),
		CodeLocation: sdktypes.GetCodeLocationCanonicalString(sdktypes.GetTriggerCodeLocation(trigger)),
	}, nil
}

func (db *gormdb) CreateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	return db.transaction(ctx, func(tx *tx) error {
		t, err := triggerToRecord(ctx, tx, trigger)
		if err != nil {
			return err
		}

		if err := tx.db.Create(t).Error; err != nil {
			return translateError(err)
		}
		return nil
	})
}

func (db *gormdb) UpdateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	return db.transaction(ctx, func(tx *tx) error {
		curr, err := tx.GetTrigger(ctx, sdktypes.GetTriggerID(trigger))
		if err != nil {
			return err
		}

		if envID := sdktypes.GetTriggerEnvID(trigger); envID != nil && sdktypes.GetTriggerEnvID(curr).String() != envID.String() {
			return sdkerrors.ErrConflict
		}

		if connID := sdktypes.GetTriggerConnectionID(trigger); connID != nil && sdktypes.GetTriggerConnectionID(curr).String() != connID.String() {
			return sdkerrors.ErrConflict
		}

		t, err := triggerToRecord(ctx, tx, trigger)
		if err != nil {
			return err
		}

		if err := tx.db.Updates(t).Error; err != nil {
			return translateError(err)
		}
		return nil
	})
}

func (db *gormdb) GetTrigger(ctx context.Context, id sdktypes.TriggerID) (sdktypes.Trigger, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseTrigger, "trigger_id = ?", id.String())
}

func (db *gormdb) DeleteTrigger(ctx context.Context, id sdktypes.TriggerID) error {
	var m scheme.Trigger
	if err := db.db.WithContext(ctx).Where("trigger_id = ?", id.String()).Delete(&m).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) ListTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	q := db.db.WithContext(ctx)
	if filter.EnvID != nil {
		q = q.Where("env_id = ?", filter.EnvID.String())
	}

	if filter.ConnectionID != nil {
		q = q.Where("connection_id = ?", filter.ConnectionID.String())
	}

	if filter.ProjectID != nil {
		q = q.Where("project_id = ?", filter.ProjectID.String())
	}

	var es []scheme.Trigger
	if err := q.Find(&es).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(es, scheme.ParseTrigger)
}
