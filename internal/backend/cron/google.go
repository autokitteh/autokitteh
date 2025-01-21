package crontab

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(INT-190): Move this to the Google integration package.

func (cr *Cron) renewGoogleCalendarEventWatchesWorkflow(wctx workflow.Context) error {
	return nil // TODO(INT-184): Implement.
}

func (cr *Cron) renewGoogleDriveEventWatchesWorkflow(wctx workflow.Context) error {
	return nil // TODO(INT-184): Implement.
}

func (cr *Cron) listGoogleCalendarConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	return nil, nil // TODO(INT-184): Implement.
}

func (cr *Cron) listGoogleDriveConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	return nil, nil // INT-184: Implement.
}

func (cr *Cron) renewGoogleCalendarEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return nil // TODO(INT-184): Implement.
}

func (cr *Cron) renewGoogleDriveEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return nil // TODO(INT-184): Implement.
}
