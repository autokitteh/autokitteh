// Package crontab is a Temporal worker that implements recurring internal
// maintenance tasks (e.g. cleanups, 3rd-party event watch renewals, etc.).
//
// In a normal state, Temporal will have: 1 available worker PER connected
// AutoKitteh server, but ONLY 1 named schedule that triggers the main
// workflow, which runs a few child workflows (1 per task), which run
// multiple activities (1 per relevant connection).
//
// Child workflows run in parallel, so they don't block each other. However,
// activities run sequentially per child workflow, so they won't overwhelm
// 3rd-party services and constrained resources that they access.
//
// Note that even in the worst case (a child workflow fails or gets terminated
// by the next invocation), successful activities won't need to run again, so
// congestions will be resolved eventually no matter what.
package crontab

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	taskQueueName = "internal-maintenance-task-queue"
	workflowName  = "internal_maintenance"
	scheduleID    = "internal_maintenance"
)

type Config struct {
	Worker   temporalclient.WorkerConfig   `koanf:"worker"`
	Workflow temporalclient.WorkflowConfig `koanf:"workflow"`
	Activity temporalclient.ActivityConfig `koanf:"activity"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{},
	}

	scheduleSpec = client.ScheduleSpec{
		Intervals: []client.ScheduleIntervalSpec{{Every: 8 * time.Hour}},
		Jitter:    10 * time.Minute, // Avoid on-the-hour spikes.
	}
	scheduleAction = &client.ScheduleWorkflowAction{
		TaskQueue: taskQueueName,
		Workflow:  workflowName,
	}
	schedulePolicies = &client.SchedulePolicies{
		// https://pkg.go.dev/go.temporal.io/api/enums/v1#ScheduleOverlapPolicy
		Overlap: enums.SCHEDULE_OVERLAP_POLICY_TERMINATE_OTHER,
	}
)

type Crontab struct {
	cfg         *Config
	logger      *zap.Logger
	temporal    temporalclient.Client
	connections sdkservices.Connections
	vars        sdkservices.Vars
	oauth       sdkservices.OAuth
}

func New(c *Config, l *zap.Logger, t temporalclient.Client) *Crontab {
	return &Crontab{cfg: c, logger: l, temporal: t}
}

func (ct *Crontab) Start(ctx context.Context, c sdkservices.Connections, v sdkservices.Vars, o sdkservices.OAuth) error {
	ct.connections = c
	ct.vars = v
	ct.oauth = o

	// Configure a worker for the internal maintenance workflow.
	w := temporalclient.NewWorker(ct.logger, ct.temporal.Temporal(), taskQueueName, ct.cfg.Worker)
	if w == nil {
		return nil
	}

	w.RegisterWorkflowWithOptions(ct.workflow, workflow.RegisterOptions{Name: workflowName})

	w.RegisterWorkflow(ct.renewGoogleCalendarEventWatchesWorkflow)
	w.RegisterActivity(ct.listGoogleCalendarConnectionsActivity)
	w.RegisterActivity(ct.renewGoogleCalendarEventWatchActivity)

	w.RegisterWorkflow(ct.renewGoogleDriveEventWatchesWorkflow)
	w.RegisterActivity(ct.listGoogleDriveConnectionsActivity)
	w.RegisterActivity(ct.renewGoogleDriveEventWatchActivity)

	w.RegisterWorkflow(ct.renewJiraEventWatchesWorkflow)
	w.RegisterActivity(ct.listJiraConnectionsActivity)
	w.RegisterActivity(ct.renewJiraEventWatchActivity)

	// Start the worker.
	if err := w.Start(); err != nil {
		return fmt.Errorf("cron: start worker: %w", err)
	}

	// Create or update the internal maintenance schedule.
	var err error
	if handle, ok := ct.scheduleAlreadyCreated(ctx); ok {
		err = ct.updateSchedule(ctx, handle)
	} else {
		err = ct.createSchedule(ctx)
	}
	return err
}

func (ct *Crontab) scheduleAlreadyCreated(ctx context.Context) (client.ScheduleHandle, bool) {
	h := ct.temporal.Temporal().ScheduleClient().GetHandle(ctx, scheduleID)
	_, err := h.Describe(ctx)
	return h, err == nil
}

func (ct *Crontab) createSchedule(ctx context.Context) error {
	handle, err := ct.temporal.Temporal().ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID:      scheduleID,
		Spec:    scheduleSpec,
		Action:  scheduleAction,
		Overlap: schedulePolicies.Overlap,
	})
	if err != nil {
		return fmt.Errorf("cron: create schedule: %w", err)
	}

	ct.logger.Info("created internal maintenance schedule",
		zap.String("schedule_id", scheduleID),
		zap.String("handle_id", handle.GetID()),
	)
	return nil
}

func (ct *Crontab) updateSchedule(ctx context.Context, handle client.ScheduleHandle) error {
	ct.logger.Debug("found existing internal maintenance schedule")

	// Attention: "handle.Update" calls are susceptible to race conditions,
	// but that's not a concern here because the schedule is supposed to be
	// the same for all workers. Even if it isn't due to a partial server
	// upgrades, it will be eventually consistent when the rollout is done.
	err := handle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(spec client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			return &client.ScheduleUpdate{
				Schedule: &client.Schedule{
					Spec:   &scheduleSpec,
					Action: scheduleAction,
					Policy: schedulePolicies,
				},
			}, nil
		},
	})
	if err != nil {
		ct.logger.Warn("failed to update internal maintenance schedule", zap.Error(err))
	}

	return nil
}

// workflow is the main entry point for the internal maintenance workflow.
// It is triggered by the schedule created in [createSchedule].
func (ct *Crontab) workflow(wctx workflow.Context) error {
	// Start multiple child workflows in parallel.
	cwfs := []workflow.ChildWorkflowFuture{}
	cwfs = append(cwfs, workflow.ExecuteChildWorkflow(wctx, ct.renewGoogleCalendarEventWatchesWorkflow))
	cwfs = append(cwfs, workflow.ExecuteChildWorkflow(wctx, ct.renewGoogleDriveEventWatchesWorkflow))
	cwfs = append(cwfs, workflow.ExecuteChildWorkflow(wctx, ct.renewJiraEventWatchesWorkflow))

	// Report an error if any child workflow failed.
	errs := make([]error, 0)
	for _, future := range cwfs {
		if err := future.Get(wctx, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
