package triggers

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

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
	if trigger.ID().IsValid() {
		return sdktypes.InvalidTriggerID, errors.New("trigger id already defined")
	}

	if trigger.WebhookSlug() != "" {
		return sdktypes.InvalidTriggerID, sdkerrors.NewInvalidArgumentError("webhook slug cannot be set")
	}

	trigger = trigger.WithNewID()

	sl := m.sl.With("trigger_id", trigger.ID())

	switch trigger.SourceType() {
	case sdktypes.TriggerSourceTypeWebhook:
		trigger = webhookssvc.InitTrigger(trigger)
		sl.With("slug", trigger.WebhookSlug()).Infof("creating webhook trigger with slug %q", trigger.WebhookSlug())
	case sdktypes.TriggerSourceTypeSchedule:
		if err := m.scheduler.Create(ctx, trigger.ID(), trigger.Schedule()); err != nil {
			return sdktypes.InvalidTriggerID, fmt.Errorf("create schedule: %w", err)
		}
		sl.With("schedule", trigger.Schedule()).Infof("creating schedule trigger with spec %q", trigger.Schedule())
	case sdktypes.TriggerSourceTypeConnection:
		sl.With("connection", trigger.ConnectionID()).Infof("creating connection trigger with connection %q", trigger.ConnectionID())
	default:
		return sdktypes.InvalidTriggerID, sdkerrors.NewInvalidArgumentError("unsupported source type")
	}

	if err := m.db.CreateTrigger(ctx, trigger); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	return trigger.ID(), nil
}

func (m *triggers) Update(ctx context.Context, next sdktypes.Trigger) error {
	next = next.WithWebhookSlug("") // ignore webhook slug.

	if next.Equal(sdktypes.InvalidTrigger) {
		// no update
		return nil
	}

	triggerID := next.ID()
	curr, err := m.db.GetTriggerByID(ctx, triggerID)
	if err != nil {
		return err
	}

	if next.Equal(curr) {
		// no update
		return nil
	}

	if curr.SourceType() != next.SourceType() {
		return sdkerrors.NewInvalidArgumentError("cannot update source type")
	}

	if next.SourceType() == sdktypes.TriggerSourceTypeSchedule {
		if err := m.scheduler.Update(ctx, triggerID, next.Schedule()); err != nil {
			return fmt.Errorf("update schedule: %w", err)
		}
	}

	return m.db.UpdateTrigger(ctx, next)
}

// Delete implements sdkservices.Triggers.
func (m *triggers) Delete(ctx context.Context, triggerID sdktypes.TriggerID) error {
	trigger, err := m.db.GetTriggerByID(ctx, triggerID)
	if err != nil {
		return err
	}

	if trigger.SourceType() == sdktypes.TriggerSourceTypeSchedule {
		if err := m.scheduler.Delete(ctx, triggerID); err != nil {
			return fmt.Errorf("delete schedule: %w", err)
		}
	}

	return errors.Join(err, m.db.DeleteTrigger(ctx, triggerID))
}

// Get implements sdkservices.Triggers.
func (m *triggers) Get(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	return m.db.GetTriggerByID(ctx, triggerID)
}

// List implements sdkservices.Triggers.
func (m *triggers) List(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	return m.db.ListTriggers(ctx, filter)
}
