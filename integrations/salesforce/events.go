package salesforce

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (h handler) dispatchEvent(payload map[string]any, eventType string) {
	akEvent, err := h.transformEvent(payload, eventType)
	if err != nil {
		h.logger.Error("failed to transform event", zap.Error(err))
		return
	}

	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, instanceURLVar, "")
	if err != nil {
		h.logger.Error("failed to find connection IDs", zap.Error(err))
		return
	}

	h.dispatchAsyncEventsToConnections(cids, akEvent)
}

// transformEvent transforms the received Salesforce event payload into an AutoKitteh event.
func (h handler) transformEvent(salesforceEvent any, eventType string) (sdktypes.Event, error) {
	l := h.logger.With(
		zap.String("event_type", eventType),
		zap.Any("event", salesforceEvent),
	)

	wrapped, err := sdktypes.WrapValue(salesforceEvent)
	if err != nil {
		l.Error("failed to wrap Salesforce event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("failed to convert wrapped Salesforce event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("failed to convert protocol buffer to SDK event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}

func (h handler) dispatchAsyncEventsToConnections(cids []sdktypes.ConnectionID, e sdktypes.Event) {
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	for _, cid := range cids {
		eid, err := h.dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		l := h.logger.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("event dispatched")
	}
}
