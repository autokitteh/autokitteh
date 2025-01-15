package drive

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/api/drive/v3"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// ConstructEvents returns all events from a single notification
func ConstructEvents(ctx context.Context, vars sdkservices.Vars, cids []sdktypes.ConnectionID) ([]sdktypes.Event, error) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	if len(cids) == 0 {
		return nil, nil
	}

	a := api{logger: l, vars: vars, cid: cids[0]}
	changes, err := a.listChanges(ctx)
	if err != nil {
		l.Error("Failed to list Google Drive events", zap.Error(err))
		return nil, err
	}

	if len(changes) == 0 {
		return nil, nil
	}

	var events []sdktypes.Event
	for _, change := range changes {
		event, err := constructSingleEvent(a, change)
		if err != nil {
			l.Error("Failed to construct event", zap.Error(err))
			continue
		}
		if event.IsValid() {
			events = append(events, event)
		}
	}

	return events, nil
}

// constructSingleEvent handles a single change event
func constructSingleEvent(a api, change *drive.Change) (sdktypes.Event, error) {
	l := a.logger

	// Convert the raw data to an AutoKitteh event
	wrapped, err := sdktypes.WrapValue(change)
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
		EventType: "change",
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}
