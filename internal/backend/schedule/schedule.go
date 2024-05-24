package schedule

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type temporalSchedule struct {
	tmprl  temporalclient.Client
	events sdkservices.Events
	z      *zap.Logger
}

func New(z *zap.Logger, events sdkservices.Events, tc temporalclient.Client) sdkservices.Scheduler {
	return &temporalSchedule{z: z, events: events, tmprl: tc}
}

func (tsc *temporalSchedule) newScheduleEvent(ctx context.Context) (sdktypes.EventID, error) {
	event := kittehs.Must1(sdktypes.EventFromProto(&sdktypes.EventPB{EventType: sdktypes.SchedulerEventTriggerType}))
	return tsc.events.Save(ctx, event)
}

func (tsc *temporalSchedule) Create(ctx context.Context, scheduleID string, schedule string, triggerID sdktypes.TriggerID) error {
	eventID, err := tsc.newScheduleEvent(ctx)
	if err != nil {
		return fmt.Errorf("shdedule: create event: %w", err)
	}

	z := tsc.z.With(zap.String("event_id", eventID.String())).With(zap.String("schedule_id", scheduleID))
	z.Debug("create schedule event", zap.String("schedule", schedule))

	_, err = tsc.tmprl.Temporal().ScheduleClient().Create(ctx,
		client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{schedule},
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        eventID.String(), // workflowID
				Workflow:  sdktypes.SchedulerWorkflow,
				TaskQueue: sdktypes.ScheduleTaskQueueName,
				Args:      []interface{}{scheduleWorkflowInput{EventID: eventID, TriggerID: triggerID}},
			},
		})
	if err != nil {
		z.Error("Failed creating schedule workflow, orphaned event", zap.Error(err))
		return fmt.Errorf("schedule: create schedule workflow: %w", err)
	}
	z.Info("created schedule workflow")
	return nil
}

func (tsc *temporalSchedule) Delete(ctx context.Context, scheduleID string) error {
	if scheduleID == "" {
		tsc.z.Error("Failed deleting schedule workflow. No scheduleID found")
		return fmt.Errorf("schedule: delete schedule workflow: no scheduleID")
	}
	z := tsc.z.With(zap.String("schedule_id", scheduleID))

	scheduleHandle := tsc.tmprl.Temporal().ScheduleClient().GetHandle(ctx, scheduleID) // validity of scheduleID is not checked by temporal
	if err := scheduleHandle.Delete(ctx); err != nil {
		z.Error("Failed deleting schedule workflow", zap.Error(err))
		return fmt.Errorf("schedule: delete scheduler workflow: %w", err)
	}
	z.Info("deleted schedule workflow")
	// REVIEW: create and save cancel schedule event?
	return nil
}

func (tsc *temporalSchedule) Update(ctx context.Context, scheduleID string, scheduleStr string) error {
	if scheduleID == "" {
		tsc.z.Error("Failed updating schedule workflow. No scheduleID found")
		return fmt.Errorf("schedule: upadate schedule workflow: no scheduleID")
	}
	z := tsc.z.With(zap.String("schedule_id", scheduleID))

	scheduleHandle := tsc.tmprl.Temporal().ScheduleClient().GetHandle(ctx, scheduleID) // validity of scheduleID is not checked by temporal
	err := scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule.Description.Schedule.Spec = &client.ScheduleSpec{CronExpressions: []string{scheduleStr}}
			return &client.ScheduleUpdate{Schedule: &schedule.Description.Schedule}, nil
		},
	})
	if err != nil {
		z.Error("Failed updating schedule workflow", zap.Error(err))
		return fmt.Errorf("schedule: update scheduler workflow: %w", err)
	}
	z.Info("updated schedule workflow")
	return nil
}
