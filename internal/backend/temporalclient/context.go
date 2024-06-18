package temporalclient

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

type contextKey string

var workflowContextKey = contextKey("autokitteh_workflow_context")

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
func (wctx *WorkflowContextAsGoContext) Value(key any) any {
	if key == workflowContextKey {
		return wctx.Context
	}

	return wctx.Context.Value(key)
}

func (wctx *WorkflowContextAsGoContext) Done() <-chan struct{} { return wctx.done }

func GetWorkflowContext(ctx context.Context) workflow.Context {
	if wctx, _ := ctx.(*WorkflowContextAsGoContext); wctx != nil {
		return wctx.Context
	}

	return nil
}
