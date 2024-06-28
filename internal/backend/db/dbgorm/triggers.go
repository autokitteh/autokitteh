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
	"gorm.io/gorm"
)

func (gdb *gormdb) withUserTriggers(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "trigger")
}

func (gdb *gormdb) createTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	idsToVerify := []*sdktypes.UUID{&trigger.ProjectID, &trigger.EnvID}
	if trigger.ConnectionID != sdktypes.BuiltinSchedulerConnectionID.UUIDValue() {
		idsToVerify = append(idsToVerify, &trigger.ConnectionID)
	}
	createFunc := func(tx *gorm.DB, user *scheme.User) error { return tx.Create(trigger).Error }
	return gdb.createEntityWithOwnership(ctx, createFunc, trigger, idsToVerify...)
}

func (gdb *gormdb) deleteTrigger(ctx context.Context, triggerID sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isUserEntity(ctx, triggerID); err != nil {
			return err
		}
		return tx.db.Delete(&scheme.Trigger{TriggerID: triggerID}).Error
	})
}

func (gdb *gormdb) updateTrigger(ctx context.Context, triggerID sdktypes.UUID, data map[string]any) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isUserEntity(ctx, triggerID); err != nil {
			return err
		}
		allowedFields := []string{"ConnectionID", "EventType", "Filter", "CodeLocation", "Data"}
		return tx.db.Model(&scheme.Trigger{TriggerID: triggerID}).Select(allowedFields).Updates(data).Error
	})
}

func (gdb *gormdb) getTrigger(ctx context.Context, triggerID sdktypes.UUID) (*scheme.Trigger, error) {
	return getOne[scheme.Trigger](gdb.withUserTriggers(ctx), "trigger_id = ?", triggerID)
}

func (gdb *gormdb) listTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]scheme.Trigger, error) {
	q := gdb.withUserTriggers(ctx)

	if filter.EnvID.IsValid() {
		q = q.Where("env_id = ?", filter.EnvID.UUIDValue())
	}

	if filter.ConnectionID.IsValid() {
		q = q.Where("connection_id = ?", filter.ConnectionID.UUIDValue())
	}

	if filter.ProjectID.IsValid() {
		q = q.Where("project_id = ?", filter.ProjectID.UUIDValue())
	}

	var ts []scheme.Trigger
	if err := q.Find(&ts).Error; err != nil {
		return nil, err
	}
	return ts, nil
}

func (db *gormdb) triggerToRecord(ctx context.Context, trigger sdktypes.Trigger) (*scheme.Trigger, error) {
	connID := trigger.ConnectionID()

	conn, err := db.GetConnection(ctx, connID)
	if err != nil {
		return nil, fmt.Errorf("get trigger connection: %w", err)
	}

	projID := conn.ProjectID()
	envID := trigger.EnvID()
	if envID.IsValid() {
		env, err := db.GetEnvByID(ctx, envID)
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

func (db *gormdb) CreateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	if err := trigger.Strict(); err != nil {
		return err
	}

	// Note: building trigger record involves non-transactionl fetching connectionID and projectID from the DB
	t, err := db.triggerToRecord(ctx, trigger)
	if err != nil {
		return err
	}

	return translateError(db.createTrigger(ctx, t))
}

func (db *gormdb) DeleteTrigger(ctx context.Context, id sdktypes.TriggerID) error {
	return translateError(db.deleteTrigger(ctx, id.UUIDValue()))
}

func (db *gormdb) UpdateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	curr, err := db.GetTrigger(ctx, trigger.ID())
	if err != nil {
		return err
	}

	if envID := trigger.EnvID(); envID.IsValid() && curr.EnvID() != envID {
		return sdkerrors.ErrConflict
	}

	// Note: building trigger record involves non-transactionl fetching connectionID and projectID from the DB
	t, err := db.triggerToRecord(ctx, trigger)
	if err != nil {
		return err
	}
	// update and set null fields as well, e.g. nullify connection, data etc..
	updateData := map[string]any{
		"ConnectionID": t.ConnectionID,
		"EventType":    t.EventType,
		"Filter":       t.Filter,
		"CodeLocation": t.CodeLocation,
		"Data":         t.Data,
	}
	return translateError(db.updateTrigger(ctx, trigger.ID().UUIDValue(), updateData))
}

func (db *gormdb) GetTrigger(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	t, err := db.getTrigger(ctx, triggerID.UUIDValue())
	if t == nil || err != nil {
		return sdktypes.InvalidTrigger, translateError(err)
	}
	return scheme.ParseTrigger(*t)
}

func (db *gormdb) ListTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	ts, err := db.listTriggers(ctx, filter)
	if ts == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(ts, scheme.ParseTrigger)
}
