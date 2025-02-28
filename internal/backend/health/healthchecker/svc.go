package healthchecker

import (
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/health/healthreporter"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type healthChecker struct {
	db     db.DB
	l      *zap.Logger
	tc     temporalclient.Client
	checks map[string]healthreporter.HealthReporter
}

type Deps struct {
	fx.In

	L *zap.Logger

	DB       db.DB                 `optional:"true"`
	Temporal temporalclient.Client `optional:"true"`
}

func New(deps Deps) healthreporter.HealthReporter {
	checker := &healthChecker{db: deps.DB, l: deps.L, tc: deps.Temporal, checks: map[string]healthreporter.HealthReporter{}}

	if deps.DB != nil {
		checker.checks["db"] = deps.DB
	}

	if deps.Temporal != nil {
		checker.checks["temporal"] = deps.Temporal
	}

	return checker
}

func (h *healthChecker) Report() error {
	allHealthErrors := []error{}
	for name, reporter := range h.checks {
		if err := reporter.Report(); err != nil {
			h.l.Error("health check error", zap.String("service", name), zap.Error(err))
			allHealthErrors = append(allHealthErrors, err)
		}
	}

	return errors.Join(allHealthErrors...)
}
