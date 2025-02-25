package modules

import (
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/testtools"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func IsAKModuleExecutorID(xid sdktypes.ExecutorID) bool {
	switch xid {
	case fixtures.ModuleExecutorID, testtools.ExecutorID:
		return true
	default:
		return false
	}
}
