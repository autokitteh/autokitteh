package fixtures

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	SchedulerEventTriggerType = "scheduler"
	SchedulerTickEventType    = "schedule_tick"
	ScheduleExpression        = "schedule"
	SchedulerConnectionName   = "cron"
)

var (
	CallOptsCtorSymbol = sdktypes.NewSymbol("callopts")
	TimeoutError       = sdktypes.NewSymbolValue(sdktypes.NewSymbol("timeout"))

	// The internal session module.
	ModuleExecutorID = sdktypes.NewExecutorID(NewBuiltinIntegrationID("ak"))
)

func NewBuiltinIntegrationID(name string) sdktypes.IntegrationID {
	return sdktypes.NewIntegrationIDFromName(name)
}

func NewBuiltinIntegrationExecutorID(name string) sdktypes.ExecutorID {
	return sdktypes.NewExecutorID(NewBuiltinIntegrationID(name))
}

func NewBuiltinExecutor(xid sdktypes.ExecutorID, opts ...sdkmodule.Optfn) sdkexecutor.Executor {
	mod := sdkmodule.New(opts...)
	vs := kittehs.Must1(mod.Configure(context.TODO(), xid, sdktypes.InvalidConnectionID))
	return sdkexecutor.NewExecutor(mod, xid, vs)
}
