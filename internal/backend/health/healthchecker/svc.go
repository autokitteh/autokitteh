package healthchecker

import (
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/health/healthreporter"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.uber.org/zap"
)

type healthChecker struct {
	db     db.DB
	z      *zap.Logger
	tc     temporalclient.Client
	checks map[string]healthreporter.HealthReporter
}

func New(db db.DB, z *zap.Logger, tc temporalclient.Client) healthreporter.HealthReporter {
	checker := &healthChecker{db: db, z: z, tc: tc, checks: map[string]healthreporter.HealthReporter{}}

	checker.checks["db"] = db
	checker.checks["temporal"] = tc
	return checker
}

func (h *healthChecker) Report() []error {
	allHealthErrors := []error{}
	for name, reporter := range h.checks {
		if reporterErrors := reporter.Report(); len(reporterErrors) > 0 {
			h.z.Error("health check error", zap.Errors(name, reporterErrors))
			allHealthErrors = append(allHealthErrors, reporterErrors...)
		}
	}

	return allHealthErrors
}
