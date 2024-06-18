package akmodules

import (
	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/ak"
	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/http"
	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/os"
	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/store"
	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/time"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func IsAKModuleExecutorID(xid sdktypes.ExecutorID) bool {
	switch xid {
	case store.ExecutorID, ak.ExecutorID, time.ExecutorID, os.ExecutorID, http.ExecutorID:
		return true
	default:
		return false
	}
}
