package temporalschedule

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/dispatcher"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type temporalScheduleImpl struct {
	z      *zap.Logger
	tmprl  temporalclient.Client
	events sdkservices.Events
}

type TemporalSchedule interface {
	CreateScheduledWorkflow(ctx context.Context, scheduleID string, schedule string, triggerID sdktypes.TriggerID) error
	DeleteSchedulerWorkflow(ctx context.Context, scheduleID string) error
	UpdateSchedulerWorkflow(ctx context.Context, scheduleID string, schedule string) error
}

func New(z *zap.Logger, events sdkservices.Events, tc temporalclient.Client) TemporalSchedule {
	return &temporalScheduleImpl{z: z, events: events, tmprl: tc}
}

func (tsc *temporalScheduleImpl) CreateScheduledWorkflow(ctx context.Context, scheduleID string, schedule string, triggerID sdktypes.TriggerID) error {
	event := kittehs.Must1(sdktypes.EventFromProto(&sdktypes.EventPB{EventType: sdktypes.SchedulerEventTriggerType}))
	eventID, err := tsc.events.Save(ctx, event)
	if err != nil {
		return fmt.Errorf("create scheduler workflow: save event: %w", err)
	}

	z := tsc.z.With(zap.String("event_id", eventID.String()))
	z.Debug("create scheduled event", zap.String("schedule", schedule))

	_, err = tsc.tmprl.ScheduleClient().Create(ctx,
		client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{schedule},
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        eventID.String(), // workflowID
				Workflow:  sdktypes.EventsWorkflow,
				TaskQueue: sdktypes.TaskQueueName,
				Args:      []interface{}{dispatcher.EventsWorkflowInput{EventID: eventID, TriggerID: &triggerID}},
			},
		})
	if err != nil {
		z.Error("Failed starting scheduler workflow, orphaned event", zap.Error(err))
		return fmt.Errorf("failed to create scheduler workflow: %w", err)
	}
	z.Debug("created scheduler workflow", zap.String("schedule_id", scheduleID))
	return nil
}

func (tsc *temporalScheduleImpl) DeleteSchedulerWorkflow(ctx context.Context, scheduleID string) error {
	if scheduleID == "" {
		tsc.z.Error("Failed delete scheduler workflow. No scheduleID found")
		return fmt.Errorf("delete scheduler workflow: no scheduleID")
	}

	scheduleHandle := tsc.tmprl.ScheduleClient().GetHandle(ctx, scheduleID) // validity of scheduleID is not checked by temporal
	if err := scheduleHandle.Delete(ctx); err != nil {
		return fmt.Errorf("delete scheduler workflow: %w", err)
	}
	// REVIEW: create and save cancel schedule event?
	return nil
}

func (tsc *temporalScheduleImpl) UpdateSchedulerWorkflow(ctx context.Context, scheduleID string, scheduleStr string) error {
	if scheduleID == "" {
		tsc.z.Error("Failed update scheduler workflow. No scheduleID found")
		return fmt.Errorf("delete scheduler workflow: no scheduleID")
	}
	z := tsc.z.With(zap.String("schedule_id", scheduleID))

	scheduleHandle := tsc.tmprl.ScheduleClient().GetHandle(ctx, scheduleID) // validity of scheduleID is not checked by temporal
	err := scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule.Description.Schedule.Spec = &client.ScheduleSpec{CronExpressions: []string{scheduleStr}}
			return &client.ScheduleUpdate{Schedule: &schedule.Description.Schedule}, nil
		},
	})
	if err != nil {
		z.Error("Failed updating scheduler workflow, orphaned event", zap.Error(err))
		return fmt.Errorf("failed to update scheduler workflow: %w", err)
	}
	z.Debug("updated scheduler workflow", zap.String("schedule", scheduleStr))
	return nil
}
