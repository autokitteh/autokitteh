package triggers

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/dispatcher"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type triggers struct {
	z      *zap.Logger
	db     db.DB
	tmprl  temporalclient.Client
	events sdkservices.Events
}

func New(z *zap.Logger, db db.DB, t temporalclient.Client, e sdkservices.Events) sdkservices.Triggers {
	return &triggers{db: db, z: z, tmprl: t, events: e}
}

func (m *triggers) Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error) {
	if trigger.ID().IsValid() {
		return sdktypes.InvalidTriggerID, errors.New("trigger id already defined")
	}
	trigger = trigger.WithNewID()

	var err error
	if schedule, found := trigger.Data()[sdktypes.ScheduleDataSection]; found && schedule.IsValid() {
		trigger, err = m.createScheduledWorkflow(ctx, trigger, schedule)
	}

	if err == nil {
		err = m.db.CreateTrigger(ctx, trigger)
	}

	if err != nil {
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
	data := trigger.Data()
	schedule, isSchedulerTrigger := data[sdktypes.ScheduleDataSection]
	prevSchedule, isSchedulerPrevTrigger := prevData[sdktypes.ScheduleDataSection]

	if !isSchedulerPrevTrigger && isSchedulerTrigger { // trigger -> scheduler trigger
		trigger, err = m.createScheduledWorkflow(ctx, trigger, schedule)
	} else if isSchedulerPrevTrigger && !isSchedulerTrigger { // scheduler trigger -> trigger
		err = m.deleteSchedulerWorkflow(ctx, data)
	} else if isSchedulerPrevTrigger && isSchedulerTrigger { // scheduler trigger -> scheduler trigger
		if schedule.String() != prevSchedule.String() { // schedule changed
			trigger, err = m.updateSchedulerWorkflow(ctx, trigger, prevData, schedule.String())
		}
	}

	// REVIEW: should we update trigger if scheduler workflow op failed?
	return errors.Join(err, m.db.UpdateTrigger(ctx, trigger))
}

// Delete implements sdkservices.Triggers.
func (m *triggers) Delete(ctx context.Context, triggerID sdktypes.TriggerID) error {
	trigger, err := m.db.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}

	data := trigger.Data()
	if schedule, found := data[sdktypes.ScheduleDataSection]; found && schedule.IsValid() {
		err = m.deleteSchedulerWorkflow(ctx, data)
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

func (m *triggers) createScheduledWorkflow(ctx context.Context, trigger sdktypes.Trigger, schedule sdktypes.Value) (sdktypes.Trigger, error) {
	scheduleStr, err := schedule.ToString()
	if err != nil || scheduleStr == "" {
		return trigger, fmt.Errorf("failed to parse trigger schedule <%v>: %w", schedule, err)
	}

	event := kittehs.Must1(sdktypes.EventFromProto(&sdktypes.EventPB{EventType: "scheduler"}))
	eventID, err := m.events.Save(ctx, event)
	if err != nil {
		return trigger, fmt.Errorf("create scheduler workflow: save event: %w", err)
	}

	z := m.z.With(zap.String("event_id", eventID.String())).With(zap.String("schedule", scheduleStr))
	z.Debug("create scheduled event", zap.String("schedule", scheduleStr))

	scheduleID := fmt.Sprintf("%s_%s", trigger.Name().String(), trigger.ID().String())
	triggerID := trigger.ID()
	_, err = m.tmprl.ScheduleClient().Create(ctx,
		client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{scheduleStr},
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        eventID.String(), // workflowID
				Workflow:  "events_workflow",
				TaskQueue: "events-task-queue",
				Args: []interface{}{
					dispatcher.EventsWorkflowInput{EventID: eventID, TriggerID: &triggerID},
				},
			},
		})
	if err != nil {
		m.z.Error("Failed starting scheduler workflow, orphaned event", zap.Error(err))
		return trigger, fmt.Errorf("failed to create scheduler workflow: %w", err)
	}
	z.Debug("created scheduler workflow", zap.String("schedule_id", scheduleID))

	// update trigger with scheduleID
	trigger = trigger.WithUpdatedData("schedule_id", sdktypes.NewStringValue(scheduleID))
	return trigger, nil
}

func extractScheduleID(data map[string]sdktypes.Value) *string {
	scheduleIDVal, found := data["schedule_id"]
	scheduleID, err := scheduleIDVal.ToString()

	if !found || err != nil || scheduleID == "" {
		return nil
	}
	return &scheduleID
}

func (m *triggers) deleteSchedulerWorkflow(ctx context.Context, data map[string]sdktypes.Value) error {
	scheduleID := extractScheduleID(data)
	if scheduleID == nil {
		m.z.Error("Failed delete scheduler workflow. No scheduleID found")
		return fmt.Errorf("delete scheduler workflow: no scheduleID")
	}

	scheduleHandle := m.tmprl.ScheduleClient().GetHandle(ctx, *scheduleID) // validity of scheduleID is not checked by temporal
	if err := scheduleHandle.Delete(ctx); err != nil {
		return fmt.Errorf("delete scheduler workflow: %w", err)
	}
	// REVIEW: create and save cancel schedule event?
	return nil
}

func (m *triggers) updateSchedulerWorkflow(
	ctx context.Context, trigger sdktypes.Trigger, prevData map[string]sdktypes.Value, scheduleStr string,
) (sdktypes.Trigger, error) {
	scheduleID := extractScheduleID(prevData)
	if scheduleID == nil {
		m.z.Error("Failed update scheduler workflow. No scheduleID found")
		return trigger, fmt.Errorf("delete scheduler workflow: no scheduleID")
	}
	z := m.z.With(zap.String("schedule_id", *scheduleID))

	trigger = trigger.WithUpdatedData("schedule_id", sdktypes.NewStringValue(*scheduleID))

	scheduleHandle := m.tmprl.ScheduleClient().GetHandle(ctx, *scheduleID) // validity of scheduleID is not checked by temporal
	err := scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule.Description.Schedule.Spec = &client.ScheduleSpec{CronExpressions: []string{scheduleStr}}
			return &client.ScheduleUpdate{Schedule: &schedule.Description.Schedule}, nil
		},
	})
	if err != nil {
		m.z.Error("Failed updating scheduler workflow, orphaned event", zap.Error(err))
		return trigger, fmt.Errorf("failed to update scheduler workflow: %w", err)
	}
	z.Debug("updated scheduler workflow", zap.String("schedule", scheduleStr))
	return trigger, nil
}
