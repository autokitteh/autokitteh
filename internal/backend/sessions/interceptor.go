package sessions

import (
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
)

type FiniteStateInterceoptor struct {
	interceptor.WorkflowInboundInterceptorBase
	Next   interceptor.WorkflowInboundInterceptor
	Notify func()
}

func NewFinitieInterceptor(notify func()) interceptor.WorkflowInboundInterceptor {
	return &FiniteStateInterceoptor{
		Notify: notify,
	}
}

func (f *FiniteStateInterceoptor) ExecuteWorkflow(
	ctx workflow.Context,
	in *interceptor.ExecuteWorkflowInput) (ret interface{}, err error) {
	if !workflow.IsReplaying(ctx) {
		fmt.Println("Started")
	}
	ex, err := f.Next.ExecuteWorkflow(ctx, in)
	if !workflow.IsReplaying(ctx) {
		fmt.Println("Finished")
		f.Notify()
	}
	return ex, err
}
