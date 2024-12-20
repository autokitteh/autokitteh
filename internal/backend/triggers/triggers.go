package triggers

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/scheduler"
	"go.autokitteh.dev/autokitteh/internal/backend/webhookssvc"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type triggers struct {
	sl        *zap.SugaredLogger
	db        db.DB
	scheduler *scheduler.Scheduler
}

func New(l *zap.Logger, db db.DB, scheduler *scheduler.Scheduler) sdkservices.Triggers {
	return &triggers{db: db, sl: l.Sugar(), scheduler: scheduler}
}

func (m *triggers) Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidTriggerID,
		"write:create",
		authz.WithData("trigger", trigger),
		authz.WithAssociationWithID("connection", trigger.ConnectionID()),
		authz.WithAssociationWithID("project", trigger.ProjectID()),
	); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	if trigger.ID().IsValid() {
		return sdktypes.InvalidTriggerID, errors.New("trigger ID already defined")
	}

	if err := trigger.Strict(); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	if trigger.WebhookSlug() != "" {
		return sdktypes.InvalidTriggerID, sdkerrors.NewInvalidArgumentError("webhook slug cannot be set")
	}

	trigger = trigger.WithNewID()

	sl := m.sl.With("trigger_id", trigger.ID())

	if trigger.SourceType() == sdktypes.TriggerSourceTypeWebhook {
		trigger = webhookssvc.InitTrigger(trigger)
	}

	if err := m.db.CreateTrigger(ctx, trigger); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	switch trigger.SourceType() {
	case sdktypes.TriggerSourceTypeWebhook:
		sl.With("slug", trigger.WebhookSlug()).Infof("created webhook trigger with slug %q", trigger.WebhookSlug())
	case sdktypes.TriggerSourceTypeSchedule:
		// TODO: If this fails, we need to remove the trigger.
		if err := m.scheduler.Create(ctx, trigger.ID(), trigger.Schedule()); err != nil {
			return sdktypes.InvalidTriggerID, fmt.Errorf("create schedule: %w", err)
		}
		sl.With("schedule", trigger.Schedule()).Infof("created schedule trigger with spec %q", trigger.Schedule())
	case sdktypes.TriggerSourceTypeConnection:
		sl.With("connection", trigger.ConnectionID()).Infof("created connection trigger with connection %q", trigger.ConnectionID())
	default:
		return sdktypes.InvalidTriggerID, sdkerrors.NewInvalidArgumentError("unsupported source type")
	}

	return trigger.ID(), nil
}

func (m *triggers) Update(ctx context.Context, trigger sdktypes.Trigger) error {
	if err := authz.CheckContext(
		ctx,
		trigger.ID(),
		"update:update",
		authz.WithData("trigger", trigger),
	); err != nil {
		return err
	}

	trigger = trigger.WithWebhookSlug("") // ignore webhook slug.

	if trigger.IsZero() {
		// no update
		return nil
	}

	triggerID := trigger.ID()
	curr, err := m.db.GetTriggerByID(ctx, triggerID)
	if err != nil {
		return err
	}

	if trigger.Equal(curr) {
		// no update
		return nil
	}

	if curr.SourceType() != trigger.SourceType() {
		return sdkerrors.NewInvalidArgumentError("cannot update source type")
	}

	if err := m.db.UpdateTrigger(ctx, trigger); err != nil {
		return err
	}

	if trigger.SourceType() == sdktypes.TriggerSourceTypeSchedule {
		// TODO: if this fails, we need to revert the trigger.
		if err := m.scheduler.Update(ctx, triggerID, trigger.Schedule()); err != nil {
			return fmt.Errorf("update schedule: %w", err)
		}
	}

	return nil
}

// Delete implements sdkservices.Triggers.
func (m *triggers) Delete(ctx context.Context, triggerID sdktypes.TriggerID) error {
	if err := authz.CheckContext(ctx, triggerID, "write:delete"); err != nil {
		return err
	}

	trigger, err := m.db.GetTriggerByID(ctx, triggerID)
	if err != nil {
		return err
	}

	if err := m.db.DeleteTrigger(ctx, triggerID); err != nil {
		return fmt.Errorf("delete trigger: %w", err)
	}

	if trigger.SourceType() == sdktypes.TriggerSourceTypeSchedule {
		// If this fails, the trigger will not work, which is fine.
		if err := m.scheduler.Delete(ctx, triggerID); err != nil {
			return fmt.Errorf("delete schedule: %w", err)
		}
	}

	return nil
}

// Get implements sdkservices.Triggers.
func (m *triggers) Get(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	if err := authz.CheckContext(ctx, triggerID, "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
		return sdktypes.InvalidTrigger, err
	}

	return m.db.GetTriggerByID(ctx, triggerID)
}

// List implements sdkservices.Triggers.
func (m *triggers) List(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidTriggerID,
		"read:list",
		authz.WithData("filter", filter),
		authz.WithAssociationWithID("connection", filter.ConnectionID),
		authz.WithAssociationWithID("project", filter.ProjectID),
	); err != nil {
		return nil, err
	}

	return m.db.ListTriggers(ctx, filter)
}
