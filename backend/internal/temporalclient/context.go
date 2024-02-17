package temporalclient

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

type WorkflowContextAsGoContext struct {
	workflow.Context
	done chan struct{}
}

func NewWorkflowContextAsGOContext(ctx workflow.Context) context.Context {
	done := make(chan struct{})

	workflow.Go(ctx, func(ctx1 workflow.Context) {
		_ = ctx.Done().Receive(ctx1, nil)
		close(done)
	})

	return &WorkflowContextAsGoContext{Context: ctx, done: done}
}

func (wctx *WorkflowContextAsGoContext) Deadline() (time.Time, bool) { return wctx.Context.Deadline() }
func (wctx *WorkflowContextAsGoContext) Err() error                  { return wctx.Context.Err() }
func (wctx *WorkflowContextAsGoContext) Value(key any) any           { return wctx.Context.Value(key) }
func (wctx *WorkflowContextAsGoContext) Done() <-chan struct{}       { return wctx.done }
