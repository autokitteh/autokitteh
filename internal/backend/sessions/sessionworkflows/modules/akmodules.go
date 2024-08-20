package modules

import (
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/os"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/store"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/time"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func IsAKModuleExecutorID(xid sdktypes.ExecutorID) bool {
	switch xid {
	case store.ExecutorID, time.ExecutorID, os.ExecutorID, fixtures.ModuleExecutorID:
		return true
	default:
		return false
	}
}
