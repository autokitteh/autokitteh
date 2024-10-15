package modules

import (
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/http"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/os"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/testtools"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/time"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func IsAKModuleExecutorID(xid sdktypes.ExecutorID) bool {
	switch xid {
	case time.ExecutorID, os.ExecutorID, fixtures.ModuleExecutorID, http.ExecutorID, testtools.ExecutorID:
		return true
	default:
		return false
	}
}
