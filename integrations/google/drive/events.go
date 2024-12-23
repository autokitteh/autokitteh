package drive

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
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

	// TODO: Workaround until ENG-1612
	if len(events) == 0 {
		return sdktypes.InvalidEvent, nil
	}
	latestEvent := events[len(events)-1]

	// https://developers.google.com/drive/api/reference/rest/v3/files
	eventType := "file_updated"
	if latestEvent.Removed {
		eventType = "file_deleted"
	}

	// TODO: ENG-1820 - Handle file creation events

	// Convert the raw data to an AutoKitteh event.
	wrapped, err := sdktypes.WrapValue(latestEvent)
	if err != nil {
		l.Error("Failed to wrap Google Drive event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Google Drive event", zap.Error(err))
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
