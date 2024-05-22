package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func triggerToRecord(ctx context.Context, tx *tx, trigger sdktypes.Trigger) (*scheme.Trigger, error) {
	connID := trigger.ConnectionID()

	conn, err := tx.GetConnection(ctx, connID)
	if err != nil {
		return nil, fmt.Errorf("get trigger connection: %w", err)
	}

	projID := conn.ProjectID()
	envID := trigger.EnvID()
	if envID.IsValid() {
		env, err := tx.GetEnvByID(ctx, envID)
		if err != nil {
			return nil, fmt.Errorf("get trigger env: %w", err)
		}

		if projID.IsValid() && projID != env.ProjectID() {
			return nil, fmt.Errorf("env and connection project mismatch: %v != %v", projID, env.ProjectID())
		}
		projID = env.ProjectID()
	}

	if !projID.IsValid() {
		return nil, fmt.Errorf("cannot guess projectID from either Env or Connection")
	}

	data, err := json.Marshal(trigger.Data())
	if err != nil {
		return nil, fmt.Errorf("marshal trigger data: %w", err)
	}

	name := trigger.Name()
	uniqueName := fmt.Sprintf("%s/%s", envID.String(), name)

	return &scheme.Trigger{
		TriggerID:    trigger.ID().UUIDValue(),
		EnvID:        envID.UUIDValue(),
		ProjectID:    projID.UUIDValue(),
		ConnectionID: connID.UUIDValue(),
		EventType:    trigger.EventType(),
		Filter:       trigger.Filter(),
		CodeLocation: trigger.CodeLocation().CanonicalString(),
		Name:         trigger.Name().String(),
		Data:         data,
		UniqueName:   uniqueName,
	}, nil
}

func (db *gormdb) createTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	return db.db.WithContext(ctx).Create(trigger).Error
}

func (db *gormdb) CreateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	if err := trigger.Strict(); err != nil {
		return err
	}

	return db.transaction(ctx, func(tx *tx) error {
		t, err := triggerToRecord(ctx, tx, trigger)
		if err != nil {
			return err
		}

		return translateError(tx.createTrigger(ctx, t))
	})
}

func (db *gormdb) UpdateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	return db.transaction(ctx, func(tx *tx) error {
		curr, err := tx.GetTrigger(ctx, trigger.ID())
		if err != nil {
			return err
		}

		if envID := trigger.EnvID(); envID.IsValid() && curr.EnvID() != envID {
			return sdkerrors.ErrConflict
		}

		if connID := trigger.ConnectionID(); connID.IsValid() && curr.ConnectionID() != connID {
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
	return getOneWTransform(db.db, ctx, scheme.ParseTrigger, "trigger_id = ?", id.UUIDValue())
}

func (db *gormdb) deleteTrigger(ctx context.Context, id sdktypes.UUID) error {
	return db.db.WithContext(ctx).Delete(&scheme.Trigger{TriggerID: id}).Error
}

func (db *gormdb) DeleteTrigger(ctx context.Context, id sdktypes.TriggerID) error {
	return translateError(db.deleteTrigger(ctx, id.UUIDValue()))
}

func (db *gormdb) ListTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	q := db.db.WithContext(ctx)
	if filter.EnvID.IsValid() {
		q = q.Where("env_id = ?", filter.EnvID.UUIDValue())
	}

	if filter.ConnectionID.IsValid() {
		q = q.Where("connection_id = ?", filter.ConnectionID.UUIDValue())
	}

	if filter.ProjectID.IsValid() {
		q = q.Where("project_id = ?", filter.ProjectID.UUIDValue())
	}

	var es []scheme.Trigger
	if err := q.Find(&es).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(es, scheme.ParseTrigger)
}
