package triggers

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// NOTE(s):
// 1. atomicity. Create temporal schedule and save trigger aren't atomic (and not temporal workflow).
//    It's possible that AK will crash after creating schedule but before saving trigger. In this case
//    the schedule will be orphaned. It will continue to be triggered by temporal, but will fail to run,
//    since it needs to fetch the trigger form DB.
//    We need to decide how to handle this case
// 2. Same note on any non-atomic schedule creation and DB/trigger update

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

	if trigger.ConnectionID() == sdktypes.BuiltinSchedulerConnectionID {
		if schedule, _ := trigger.Data()[fixtures.ScheduleExpression].ToString(); schedule != "" {
			if err := m.scheduler.Create(ctx, trigger.ID().String(), schedule, trigger.ID()); err != nil {
				return sdktypes.InvalidTriggerID, err
			}
		}
	}

	// FIXME: see atomicity NOTE above
	if err := m.db.CreateTrigger(ctx, trigger); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	return trigger.ID(), nil
}

func (m *triggers) Update(ctx context.Context, trigger sdktypes.Trigger) error {
	triggerID := trigger.ID()
	prevTrigger, err := m.db.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}

	prevData := prevTrigger.Data()
	prevScheduleVal, isSchedulerPrevTrigger := prevData[fixtures.ScheduleExpression]
	isSchedulerPrevTrigger = isSchedulerPrevTrigger && prevTrigger.ConnectionID() == sdktypes.BuiltinSchedulerConnectionID

	data := trigger.Data()
	scheduleVal, isSchedulerTrigger := data[fixtures.ScheduleExpression]
	schedule, _ := scheduleVal.ToString()
	isSchedulerTrigger = isSchedulerTrigger && trigger.ConnectionID() == sdktypes.BuiltinSchedulerConnectionID

	if isSchedulerPrevTrigger {
		if !isSchedulerTrigger { // scheduler trigger -> trigger
			err = m.scheduler.Delete(ctx, triggerID.String())
		} else { // scheduler trigger -> scheduler trigger
			if schedule != prevScheduleVal.String() { // schedule changed
				err = m.scheduler.Update(ctx, triggerID.String(), schedule)
			}
		}
	} else if isSchedulerTrigger { // trigger -> scheduler trigger
		err = m.scheduler.Create(ctx, triggerID.String(), schedule, triggerID)
	}

	if err != nil {
		return err
	}
	return m.db.UpdateTrigger(ctx, trigger)
}

// Delete implements sdkservices.Triggers.
func (m *triggers) Delete(ctx context.Context, triggerID sdktypes.TriggerID) error {
	trigger, err := m.db.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}

	if trigger.ConnectionID() == sdktypes.BuiltinSchedulerConnectionID {
		err = m.scheduler.Delete(ctx, triggerID.String())
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
