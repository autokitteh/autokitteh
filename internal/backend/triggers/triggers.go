package triggers

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalschedule"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type triggers struct {
	z   *zap.Logger
	db  db.DB
	tsc temporalschedule.TemporalSchedule
}

func New(z *zap.Logger, db db.DB, tsc temporalschedule.TemporalSchedule) sdkservices.Triggers {
	return &triggers{db: db, z: z, tsc: tsc}
}

func (m *triggers) Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error) {
	if trigger.ID().IsValid() {
		return sdktypes.InvalidTriggerID, errors.New("trigger id already defined")
	}

	trigger = trigger.WithNewID()
	if schedule, _ := trigger.Data()[sdktypes.ScheduleExpression].ToString(); schedule != "" {
		scheduleID := newScheduleID(trigger)
		if err := m.tsc.CreateScheduledWorkflow(ctx, scheduleID, schedule, trigger.ID()); err != nil {
			return sdktypes.InvalidTriggerID, err
		}
		trigger = trigger.WithUpdatedData(sdktypes.ScheduleIDKey, sdktypes.NewStringValue(scheduleID))
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

	data := trigger.Data()
	scheduleVal, isSchedulerTrigger := data[sdktypes.ScheduleExpression]
	schedule, _ := scheduleVal.ToString()

	if isSchedulerPrevTrigger {
		if !isSchedulerTrigger { // scheduler trigger -> trigger
			err = m.tsc.DeleteSchedulerWorkflow(ctx, scheduleID)
			scheduleID = ""
		} else { // scheduler trigger -> scheduler trigger
			if schedule != prevScheduleVal.String() { // schedule changed
				err = m.tsc.UpdateSchedulerWorkflow(ctx, scheduleID, schedule)
			}
		}
	} else if isSchedulerTrigger { // trigger -> scheduler trigger
		scheduleID = newScheduleID(trigger)
		err = m.tsc.CreateScheduledWorkflow(ctx, scheduleID, schedule, trigger.ID())
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
	if schedule, found := data[sdktypes.ScheduleExpression]; found && schedule.IsValid() {
		scheduleID, _ := data[sdktypes.ScheduleIDKey].ToString()
		err = m.tsc.DeleteSchedulerWorkflow(ctx, scheduleID)
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
