package salesforce

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (h handler) handleSalesForceEvent(decodedEventPayload map[string]any, topicName string) {
	akEvent, err := transformEvent(h.logger, decodedEventPayload, topicName)
	if err != nil {
		h.logger.Error("failed to transform event", zap.Error(err))
		return
	}

	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, instanceURLVar, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
		return
	}

	h.dispatchAsyncEventsToConnections(cids, akEvent)
}

// Transform the received Salesforce event into an AutoKitteh event.
func transformEvent(l *zap.Logger, salesforceEvent any, eventType string) (sdktypes.Event, error) {
	l = l.With(
		zap.String("eventType", eventType),
		zap.Any("event", salesforceEvent),
	)

	wrapped, err := sdktypes.WrapValue(salesforceEvent)
	if err != nil {
		l.Error("Failed to wrap Salesforce event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Salesforce event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.Any("data", data),
			zap.Error(err),
		)
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
			l.Error("Event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("Event dispatched")
	}
}
