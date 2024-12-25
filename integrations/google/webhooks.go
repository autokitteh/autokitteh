package google

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// handler implements several AutoKitteh webhooks to save authentication
// data, as well as receive and dispatch asynchronous event notifications.
type handler struct {
	logger     *zap.Logger
	oauth      sdkservices.OAuth
	vars       sdkservices.Vars
	dispatcher sdkservices.Dispatcher
}

func NewHTTPHandler(l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars, d sdkservices.Dispatcher) handler {
	return handler{logger: l, oauth: o, vars: v, dispatcher: d}
}
