package google

import (
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// handler implements several AutoKitteh webhooks to save authentication
// data, as well as receive and dispatch asynchronous event notifications.
type handler struct {
	logger   *zap.Logger
	oauth    *oauth.OAuth
	vars     sdkservices.Vars
	dispatch sdkservices.DispatchFunc
}

func NewHTTPHandler(l *zap.Logger, o *oauth.OAuth, v sdkservices.Vars, d sdkservices.DispatchFunc) handler {
	return handler{logger: l, oauth: o, vars: v, dispatch: d}
}
