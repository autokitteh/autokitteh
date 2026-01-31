// Adapted from https://github.com/dynajoe/temporal-terraform-demo/blob/main/heartbeat/heartbeat.go.
package sessioncalls

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

func BeginHeartbeat(ctx context.Context, interval time.Duration, shouldHeartbeat func(context.Context) bool) (context.Context, func()) {
	// Create a context that can be canceled as soon as the worker is stopped
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-activity.GetWorkerStopChannel(ctx):
		case <-ctx.Done():
		}
		cancel()
	}()

	go startHeartbeats(ctx, interval, shouldHeartbeat)

	return ctx, cancel
}

func startHeartbeats(ctx context.Context, interval time.Duration, shouldHeartbeat func(context.Context) bool) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	activity.RecordHeartbeat(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if shouldHeartbeat(ctx) {
				activity.RecordHeartbeat(ctx)
			}
		}
	}
}
