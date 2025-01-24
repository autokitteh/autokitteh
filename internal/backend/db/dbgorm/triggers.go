package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) createTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	return gormErrNotFoundToForeignKey(gdb.wdb.WithContext(ctx).Create(trigger).Error)
}

func (gdb *gormdb) deleteTrigger(ctx context.Context, triggerID uuid.UUID) error {
	// NOTE: we allow delettion of triggers referenced by events. see ENG-1535
	return gdb.wdb.WithContext(ctx).Delete(&scheme.Trigger{TriggerID: triggerID}).Error
}

func (gdb *gormdb) updateTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	return gdb.wdb.WithContext(ctx).Model(&scheme.Trigger{TriggerID: trigger.TriggerID}).Updates(trigger).Error
}

func (gdb *gormdb) getTriggerByID(ctx context.Context, triggerID uuid.UUID) (*scheme.Trigger, error) {
	return getOne[scheme.Trigger](gdb.rdb.WithContext(ctx), "trigger_id = ?", triggerID)
}

func (gdb *gormdb) listTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]scheme.Trigger, error) {
	q := gdb.rdb.WithContext(ctx)

	if filter.ProjectID.IsValid() {
		q = q.Where("triggers.project_id = ?", filter.ProjectID.UUIDValue())
	}

	q = withProjectOrgID(q, filter.OrgID, "triggers")

	if filter.ConnectionID.IsValid() {
		q = q.Where("connection_id = ?", filter.ConnectionID.UUIDValue())
	}

	if !filter.SourceType.IsZero() {
		q = q.Where("source_type = ?", filter.SourceType.String())
	}

	var ts []scheme.Trigger
	if err := q.Find(&ts).Error; err != nil {
		return nil, err
	}
	return ts, nil
}

func triggerUniqueName(p string, name sdktypes.Symbol) string {
	return fmt.Sprintf("%s/%s", p, name.String())
}

func (db *gormdb) CreateTrigger(ctx context.Context, trigger sdktypes.Trigger) error {
	if err := trigger.Strict(); err != nil { // name, connection, and project
		return err
	}

	pid := trigger.ProjectID()

	uniqueName := triggerUniqueName(pid.String(), trigger.Name())

	t := &scheme.Trigger{
		Base:         based(ctx),
		ProjectID:    trigger.ProjectID().UUIDValue(),
		TriggerID:    trigger.ID().UUIDValue(),
		ConnectionID: trigger.ConnectionID().UUIDValuePtr(),
		SourceType:   trigger.SourceType().String(),
		EventType:    trigger.EventType(),
		Filter:       trigger.Filter(),
		CodeLocation: trigger.CodeLocation().CanonicalString(),
		Name:         trigger.Name().String(),
		UniqueName:   uniqueName,
		WebhookSlug:  trigger.WebhookSlug(),
		Schedule:     trigger.Schedule(),
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

	// We're not checking anything else here and modifying only the following fields.
	// This means that there'll be no error if an unmodifyable field is requested
	// to be changed by the caller, and it will not be modified in the DB.

	r.CodeLocation = trigger.CodeLocation().CanonicalString()
	r.EventType = trigger.EventType()
	r.Filter = trigger.Filter()
	r.Schedule = trigger.Schedule()
	r.Name = trigger.Name().String()
	r.UniqueName = triggerUniqueName(r.ProjectID.String(), trigger.Name())
	r.UpdatedAt = kittehs.Now().UTC()
	r.UpdatedBy = authcontext.GetAuthnUserID(ctx).UUIDValue()

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
	r, err := getOne[scheme.Trigger](db.rdb.WithContext(ctx), "webhook_slug = ?", slug)
	if err != nil {
		return sdktypes.InvalidTrigger, translateError(err)
	}

	return scheme.ParseTrigger(*r)
}
