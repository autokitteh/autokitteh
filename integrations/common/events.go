package common

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TransformEvent transforms a third-party's event payload into an AutoKitteh event.
func TransformEvent(l *zap.Logger, payload any, eventType string) (sdktypes.Event, error) {
	l = l.With(
		zap.String("event_type", eventType),
		zap.Any("payload", payload),
	)

	v, err := sdktypes.WrapValue(payload)
	if err != nil {
		l.Error("failed to wrap Linear event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	m, err := v.ToStringValuesMap()
	if err != nil {
		l.Error("failed to convert wrapped Linear event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	e, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(m, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("failed to convert protocol buffer to SDK event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	return e, nil
}

// DispatchEvent dispatches the given event to all the
// given connections, for potential asynchronous handling.
func DispatchEvent(ctx context.Context, l *zap.Logger, d sdkservices.DispatchFunc, e sdktypes.Event, cids []sdktypes.ConnectionID) {
	for _, cid := range cids {
		eid, err := d(ctx, e.WithConnectionDestinationID(cid), nil)
		l := l.With(
			zap.String("connection_id", cid.String()),
			zap.String("event_id", eid.String()),
		)
		if err != nil {
			if errors.Is(err, sdkerrors.ErrResourceExhausted) {
				l.Info("Event dispatch failed due to resource exhaustion")
			} else {
				l.Error("Event dispatch failed", zap.Error(err))
			}

			return
		}
		l.Debug("event dispatched")
	}
}
