package modules

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var ids = make(map[sdktypes.ExecutorID]bool)

func Register(id sdktypes.ExecutorID) { ids[id] = true }

func IsAKModuleExecutorID(xid sdktypes.ExecutorID) bool { return ids[xid] }
