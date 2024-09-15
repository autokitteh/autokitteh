package calendar

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func ConstructEvent(ctx context.Context, vars sdkservices.Vars, cids []sdktypes.ConnectionID) (sdktypes.Event, error) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	if len(cids) == 0 {
		return sdktypes.InvalidEvent, nil
	}

	// Enrich the event with relevant data, using the connection's sync token.
	a := api{logger: l, vars: vars, cid: cids[0]}
	events, err := a.listEvents(ctx)
	if err != nil {
		l.Error("Failed to list events", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	if len(events) != 1 {
		l.Error("Number of Google Calendar events for incoming notification != 1",
			zap.Int("numEvents", len(events)),
		)
		return sdktypes.InvalidEvent, errors.New("unexpected Google Calendar events")
	}

	// https://developers.google.com/calendar/api/v3/reference/events#resource
	eventType := "event_updated"
	if events[0].Status == "cancelled" {
		eventType = "event_deleted"
	} else if events[0].Sequence == 0 {
		eventType = "event_created"
	}

	// Convert the raw data to an AutoKitteh event.
	wrapped, err := sdktypes.WrapValue(events[0])
	if err != nil {
		l.Error("Failed to wrap Google Calendar event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Google Calendar event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}
