package dbgorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) createTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	return gormErrNotFoundToForeignKey(gdb.writer.WithContext(ctx).Create(trigger).Error)
}

func (gdb *gormdb) deleteTrigger(ctx context.Context, triggerID uuid.UUID) error {
	// NOTE: we allow delettion of triggers referenced by events. see ENG-1535
	return gdb.writer.WithContext(ctx).
		Model(&scheme.Trigger{}).
		Where("trigger_id = ?", triggerID).
		Updates(map[string]any{
			"unique_name": triggerID.String(),
			"deleted_at":  time.Now(),
		}).Error
}

func (gdb *gormdb) updateTrigger(ctx context.Context, trigger *scheme.Trigger) error {
	return gdb.writer.WithContext(ctx).Model(&scheme.Trigger{TriggerID: trigger.TriggerID}).Updates(trigger).Error
}

func (gdb *gormdb) getTriggerByID(ctx context.Context, triggerID uuid.UUID) (*scheme.Trigger, error) {
	return getOne[scheme.Trigger](gdb.reader.WithContext(ctx), "trigger_id = ?", triggerID)
}

func (gdb *gormdb) listTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]scheme.Trigger, error) {
	q := gdb.reader.WithContext(ctx)

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

	isDurable := trigger.IsDurable()
	isSync := trigger.IsSync()

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
		Timezone:     trigger.Timezone(),
		Schedule:     trigger.Schedule(),
		IsDurable:    &isDurable,
		IsSync:       &isSync,
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

	isDurable := trigger.IsDurable()
	isSync := trigger.IsSync()

	r.CodeLocation = trigger.CodeLocation().CanonicalString()
	r.EventType = trigger.EventType()
	r.Filter = trigger.Filter()
	r.Schedule = trigger.Schedule()
	r.Timezone = trigger.Timezone()
	r.Name = trigger.Name().String()
	r.UniqueName = triggerUniqueName(r.ProjectID.String(), trigger.Name())
	r.UpdatedAt = kittehs.Now().UTC()
	r.UpdatedBy = authcontext.GetAuthnUserID(ctx).UUIDValue()
	r.IsSync = &isSync
	r.IsDurable = &isDurable

	return translateError(db.updateTrigger(ctx, r))
}

func (db *gormdb) GetTriggerByID(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	r, err := db.getTriggerByID(ctx, triggerID.UUIDValue())
	if r == nil || err != nil {
		return sdktypes.InvalidTrigger, translateError(err)
	}

	return scheme.ParseTrigger(*r)
}

func (db *gormdb) GetTriggerWithActiveDeploymentByID(ctx context.Context, triggerID uuid.UUID) (sdktypes.Trigger, bool, error) {
	var triggerAndDeployment struct {
		scheme.Trigger
		HasActiveDeployment bool `gorm:"column:has_active_deployment"`
	}

	err := db.reader.WithContext(ctx).
		Model(&scheme.Trigger{}).
		Select("triggers.*, CASE WHEN deployments.deployment_id IS NOT NULL THEN true ELSE false END as has_active_deployment").
		Joins("LEFT JOIN deployments ON triggers.project_id = deployments.project_id AND deployments.state = ? AND deployments.deleted_at IS NULL",
			int32(sdktypes.DeploymentStateActive.ToProto())).
		Where("triggers.trigger_id = ?", triggerID).
		First(&triggerAndDeployment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sdktypes.InvalidTrigger, false, sdkerrors.ErrNotFound // Trigger doesn't exist.
		}
		return sdktypes.InvalidTrigger, false, translateError(err)
	}

	trigger, err := scheme.ParseTrigger(triggerAndDeployment.Trigger)
	return trigger, triggerAndDeployment.HasActiveDeployment, err
}

func (db *gormdb) ListTriggers(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	ts, err := db.listTriggers(ctx, filter)
	if ts == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(ts, scheme.ParseTrigger)
}

func (db *gormdb) GetTriggerWithActiveDeploymentByWebhookSlug(ctx context.Context, slug string) (sdktypes.Trigger, error) {
	var trigger scheme.Trigger
	err := db.reader.WithContext(ctx).
		Model(&scheme.Trigger{}).
		Joins("JOIN deployments ON triggers.project_id = deployments.project_id").
		Where("triggers.webhook_slug = ? AND deployments.state = ? AND deployments.deleted_at IS NULL",
			slug,
			int32(sdktypes.DeploymentStateActive.ToProto())).
		First(&trigger).Error
	if err != nil {
		return sdktypes.InvalidTrigger, translateError(err)
	}

	return scheme.ParseTrigger(trigger)
}
