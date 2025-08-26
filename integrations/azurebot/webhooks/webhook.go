package webhooks

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	for _, cid := range cids {
		eid, err := h.dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		l := l.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			if errors.Is(err, sdkerrors.ErrResourceExhausted) {
				l.Info("Event dispatch failed due to resource exhaustion")
			} else {
				l.Error("Event dispatch failed", zap.Error(err))
			}
			return
		}
		l.Debug("Event dispatched")
	}
}
