package temporalclient

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

var (
	workflowContextKey = new(struct{})
	spanContextKey     = new(struct{})
)

type WorkflowContextAsGoContext struct {
	workflow.Context
	spanCtx context.Context
	done    chan struct{}
}

// Creates a new GO context that gets the Done() signal from the workflow context.
// Performing long running operations in the returned context will block the workflow execution,
// which will result in Temporal's deadlock detector kicking in and kickking your butt.
func NewWorkflowContextAsGOContext(wctx workflow.Context) context.Context {
	done := make(chan struct{})

	workflow.Go(wctx, func(ctx1 workflow.Context) {
		_ = wctx.Done().Receive(ctx1, nil)
		close(done)
	})

	spanCtx, _ := wctx.Value(spanContextKey).(context.Context)

	return &WorkflowContextAsGoContext{
		Context: wctx,
		spanCtx: spanCtx,
		done:    done,
	}
}

func (wctx *WorkflowContextAsGoContext) Deadline() (time.Time, bool) { return wctx.Context.Deadline() }
func (wctx *WorkflowContextAsGoContext) Err() error                  { return wctx.Context.Err() }
func (wctx *WorkflowContextAsGoContext) Value(key any) any {
	if key == workflowContextKey {
		return true
	}

	if wctx.spanCtx != nil {
		sv := wctx.spanCtx.Value(key)
		if sv != nil {
			return sv
		}
	}

	return wctx.Context.Value(key)
}
func (wctx *WorkflowContextAsGoContext) Done() <-chan struct{} { return wctx.done }

func IsWorkflowContextAsGoContext(ctx context.Context) bool {
	// We can't just check if the ctx is *WorkflowContextAsGoContext because the user might have
	// wrapped it in another context.
	return ctx.Value(workflowContextKey) == true
}
