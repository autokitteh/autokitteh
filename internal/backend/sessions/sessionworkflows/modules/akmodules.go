package modules

import (
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows/modules/testtools"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func IsAKModuleExecutorID(xid sdktypes.ExecutorID) bool { return xid == testtools.ExecutorID }
