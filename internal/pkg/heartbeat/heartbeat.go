// Based on https://github.com/dynajoe/temporal-terraform-demo/blob/main/heartbeat/heartbeat.go.
// Reasoning: https://community.temporal.io/t/best-practices-for-long-running-activities/934/2.
package heartbeat

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

func Begin(ctx context.Context, interval time.Duration, getDetails func() []interface{}) (context.Context, func()) {
	if getDetails == nil {
		getDetails = func() []interface{} { return nil }
	}

	// Create a context that can be canceled as soon as the worker is stopped
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-activity.GetWorkerStopChannel(ctx):
		case <-ctx.Done():
		}
		cancel()
	}()

	go startHeartbeats(ctx, interval, getDetails)

	return ctx, cancel
}

func startHeartbeats(ctx context.Context, interval time.Duration, getDetails func() []interface{}) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(100 * time.Millisecond):
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	activity.RecordHeartbeat(ctx, getDetails()...)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			activity.RecordHeartbeat(ctx)
		}
	}
}
