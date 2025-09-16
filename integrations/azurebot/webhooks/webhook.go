package webhooks

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler] to
// receive, dispatch, and acknowledge asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatch      sdkservices.DispatchFunc
	integration   sdktypes.Integration
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, vars sdkservices.Vars, d sdkservices.DispatchFunc, i sdktypes.Integration) handler {
	return handler{
		logger:        l,
		vars:          vars,
		dispatch:      d,
		integration:   i,
		integrationID: i.ID(),
	}
}
