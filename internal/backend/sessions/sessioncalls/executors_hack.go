package sessioncalls

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
)

// HACK(ENG-335):
//  Pass executors for session specific calls (functions that are
//	defined in runtime script, as opposed to integrations).

var executorsForSessions = make(map[string]*sdkexecutor.Executors, 16)
