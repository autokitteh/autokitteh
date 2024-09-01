package calendar

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/api/forms/v1"

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
		EventType: "calendar_event",
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}

func lastResponse(responses []*forms.FormResponse) *forms.FormResponse {
	if len(responses) == 0 {
		return &forms.FormResponse{}
	}

	last := responses[0]
	for _, r := range responses {
		if r.LastSubmittedTime > last.LastSubmittedTime {
			last = r
		}
	}

	return last
}
