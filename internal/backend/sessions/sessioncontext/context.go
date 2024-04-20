package sessioncontext

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

type key string

const workflowContextKey key = "workflow_context"

func WithWorkflowContext(ctx context.Context, wctx workflow.Context) context.Context {
	return context.WithValue(ctx, workflowContextKey, wctx)
}

func GetWorkflowContext(ctx context.Context) workflow.Context {
	v := ctx.Value(workflowContextKey)
	if v == nil {
		return nil
	}

	return v.(workflow.Context)
}
