package dbgorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) withUserTriggers(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "trigger")
}

func (gdb *gormdb) createTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	idsToVerify := []*sdktypes.UUID{&trigger.ProjectID, &trigger.EnvID}
	createFunc := func(tx *gorm.DB, uid string) error { return tx.Create(trigger).Error }
	return gormErrNotFoundToForeignKey(
		gdb.createEntityWithOwnership(ctx, createFunc, trigger, idsToVerify...))
}

func (gdb *gormdb) deleteTrigger(ctx context.Context, triggerID sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(tx.ctx, triggerID); err != nil {
			return err
		}

		// NOTE: we allow delettion of triggeres referenced by events. see ENG-1535

		return tx.db.Delete(&scheme.Trigger{TriggerID: triggerID}).Error
	})
}

func (gdb *gormdb) updateTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(tx.ctx, trigger.TriggerID); err != nil {
			return err
		}

		return tx.db.Model(&scheme.Trigger{TriggerID: trigger.TriggerID}).Updates(trigger).Error
	})
}

func (gdb *gormdb) getTriggerByID(ctx context.Context, triggerID sdktypes.UUID) (*scheme.Trigger, error) {
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

	if !filter.SourceType.IsZero() {
		q = q.Where("source_type = ?", filter.SourceType.String())
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

func triggerUniqueName(env string, name sdktypes.Symbol) string {
	return fmt.Sprintf("%s/%s", env, name.String())
}

func (db *gormdb) triggerToRecord(ctx context.Context, trigger sdktypes.Trigger) (*scheme.Trigger, error) {
	envID := trigger.EnvID()
	env, err := db.GetEnvByID(ctx, envID)
	if err != nil {
		return nil, fmt.Errorf("get trigger env: %w", err)
	}
	projID := env.ProjectID()

	uniqueName := triggerUniqueName(envID.String(), trigger.Name())

	return &scheme.Trigger{
		TriggerID:    trigger.ID().UUIDValue(),
		EnvID:        envID.UUIDValue(),
		ProjectID:    projID.UUIDValue(),
		ConnectionID: trigger.ConnectionID().UUIDValuePtr(),
		SourceType:   trigger.SourceType().String(),
		EventType:    trigger.EventType(),
		Filter:       trigger.Filter(),
		CodeLocation: trigger.CodeLocation().CanonicalString(),
		Name:         trigger.Name().String(),
		UniqueName:   uniqueName,
		WebhookSlug:  trigger.WebhookSlug(),
		Schedule:     trigger.Schedule(),
	}, nil
}

func (db *gormdb) CreateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	if err := trigger.Strict(); err != nil { // name, connection, env but not project
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
	r, err := db.getTriggerByID(ctx, trigger.ID().UUIDValue())
	if r == nil || err != nil {
		return translateError(err)
	}

	r.CodeLocation = trigger.CodeLocation().CanonicalString()
	r.EventType = trigger.EventType()
	r.Filter = trigger.Filter()
	r.Schedule = trigger.Schedule()

	return translateError(db.updateTrigger(ctx, r))
}

func (db *gormdb) GetTriggerByID(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	r, err := db.getTriggerByID(ctx, triggerID.UUIDValue())
	if r == nil || err != nil {
		return sdktypes.InvalidTrigger, translateError(err)
	}

	return scheme.ParseTrigger(*r)
}

func (db *gormdb) ListTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	ts, err := db.listTriggers(ctx, filter)
	if ts == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(ts, scheme.ParseTrigger)
}

func (db *gormdb) GetTriggerByWebhookSlug(ctx context.Context, slug string) (sdktypes.Trigger, error) {
	r, err := getOne[scheme.Trigger](db.withUserTriggers(ctx), "webhook_slug = ?", slug)
	if err != nil {
		return sdktypes.InvalidTrigger, translateError(err)
	}

	return scheme.ParseTrigger(*r)
}
