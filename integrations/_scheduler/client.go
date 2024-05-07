package scheduler

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.NewIntegrationIDFromName("scheduler")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "scheduler",
	DisplayName:   "Scheduler (Cron)",
	Description:   "Cron-like scheduler of autokitteh events.",
	LogoUrl:       "/static/images/scheduler.svg",
	UserLinks: map[string]string{
		`1 Cron expression format ("* * * * *")`:            "https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format",
		`2 Predefined schedules and intervals ("@" format)`: "https://pkg.go.dev/github.com/robfig/cron#hdr-Predefined_schedules",
		"3 Crontab.guru - cron schedule expression editor":  "https://crontab.guru/",
	},
	ConnectionUrl: "/scheduler/connect",
}))

func New(sec sdkservices.Secrets) sdkservices.Integration {
	// No functions, only events.
	return sdkintegrations.NewIntegration(desc, sdkmodule.New())
}
