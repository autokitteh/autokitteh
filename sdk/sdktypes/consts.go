package sdktypes

import "go.autokitteh.dev/autokitteh/internal/kittehs"

const (
	SchedulerEventTriggerType = "scheduler"
	SchedulerTickEventType    = "schedule_tick"
	ScheduleExpression        = "schedule"
	SchedulerConnectionName   = "cron"
)

var BuiltinSchedulerConnectionID = kittehs.Must1(ParseConnectionID("con_3kthcr0n000000000000000000"))
