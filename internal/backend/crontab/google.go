package crontab

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(INT-190): Move this to the Google integration package.

func (ct *Crontab) renewGoogleCalendarEventWatchesWorkflow(wctx workflow.Context) error {
	return nil // TODO(INT-184): Implement.
}

func (ct *Crontab) renewGoogleDriveEventWatchesWorkflow(wctx workflow.Context) error {
	return nil // TODO(INT-184): Implement.
}

func (ct *Crontab) renewGoogleCalendarEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return nil // TODO(INT-184): Implement.
}

func (ct *Crontab) renewGoogleDriveEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return nil // TODO(INT-184): Implement.
}
