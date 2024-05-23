package triggers

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type triggers struct {
	z         *zap.Logger
	db        db.DB
	scheduler sdkservices.Scheduler
}

func New(z *zap.Logger, db db.DB, tsc sdkservices.Scheduler) sdkservices.Triggers {
	return &triggers{db: db, z: z, scheduler: tsc}
}

func (m *triggers) Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error) {
	if trigger.ID().IsValid() {
		return sdktypes.InvalidTriggerID, errors.New("trigger id already defined")
	}

	trigger = trigger.WithNewID()

	if trigger.ConnectionID() == fixtures.BuiltinSchedulerConnectionID {
		if schedule, _ := trigger.Data()[sdktypes.ScheduleExpression].ToString(); schedule != "" {
			scheduleID := newScheduleID(trigger)
			if err := m.scheduler.Create(ctx, scheduleID, schedule, trigger.ID()); err != nil {
				return sdktypes.InvalidTriggerID, err
			}
			trigger = trigger.WithUpdatedData(sdktypes.ScheduleIDKey, sdktypes.NewStringValue(scheduleID))
		}
	}

	if err := m.db.CreateTrigger(ctx, trigger); err != nil {
		return sdktypes.InvalidTriggerID, err
	}
	return trigger.ID(), nil
}

func (m *triggers) Update(ctx context.Context, trigger sdktypes.Trigger) error {
	prevTrigger, err := m.db.GetTrigger(ctx, trigger.ID())
	if err != nil {
		return err
	}

	prevData := prevTrigger.Data()
	prevScheduleVal, isSchedulerPrevTrigger := prevData[sdktypes.ScheduleExpression]
	scheduleID, _ := prevData[sdktypes.ScheduleIDKey].ToString()
	isSchedulerPrevTrigger = isSchedulerPrevTrigger && prevTrigger.ConnectionID() == fixtures.BuiltinSchedulerConnectionID

	data := trigger.Data()
	scheduleVal, isSchedulerTrigger := data[sdktypes.ScheduleExpression]
	schedule, _ := scheduleVal.ToString()
	isSchedulerTrigger = isSchedulerTrigger && trigger.ConnectionID() == fixtures.BuiltinSchedulerConnectionID

	if isSchedulerPrevTrigger {
		if !isSchedulerTrigger { // scheduler trigger -> trigger
			err = m.scheduler.Delete(ctx, scheduleID)
			scheduleID = ""
		} else { // scheduler trigger -> scheduler trigger
			if schedule != prevScheduleVal.String() { // schedule changed
				err = m.scheduler.Update(ctx, scheduleID, schedule)
			}
		}
	} else if isSchedulerTrigger { // trigger -> scheduler trigger
		scheduleID = newScheduleID(trigger)
		err = m.scheduler.Create(ctx, scheduleID, schedule, trigger.ID())
	}

	if err != nil {
		return err
	}

	if scheduleID != "" {
		trigger = trigger.WithUpdatedData(sdktypes.ScheduleIDKey, sdktypes.NewStringValue(scheduleID))
	}
	return m.db.UpdateTrigger(ctx, trigger)
}

// Delete implements sdkservices.Triggers.
func (m *triggers) Delete(ctx context.Context, triggerID sdktypes.TriggerID) error {
	trigger, err := m.db.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}

	data := trigger.Data()
	if trigger.ConnectionID() == fixtures.BuiltinSchedulerConnectionID {
		scheduleID, _ := data[sdktypes.ScheduleIDKey].ToString()
		err = m.scheduler.Delete(ctx, scheduleID)
	}
	return errors.Join(err, m.db.DeleteTrigger(ctx, triggerID))
}

// Get implements sdkservices.Triggers.
func (m *triggers) Get(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	return sdkerrors.IgnoreNotFoundErr(m.db.GetTrigger(ctx, triggerID))
}

// List implements sdkservices.Triggers.
func (m *triggers) List(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	return m.db.ListTriggers(ctx, filter)
}

func newScheduleID(trigger sdktypes.Trigger) string {
	return fmt.Sprintf("%s_%s", trigger.Name().String(), trigger.ID().String())
}
