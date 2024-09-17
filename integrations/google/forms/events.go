package forms

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/api/forms/v1"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func ConstructEvent(ctx context.Context, vars sdkservices.Vars, formsEvent map[string]any, cids []sdktypes.ConnectionID) (sdktypes.Event, error) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	if len(cids) == 0 {
		return sdktypes.InvalidEvent, nil
	}

	// Enrich the event with relevant data, using API calls.
	a := api{vars: vars, cid: cids[0]}
	switch WatchEventType(formsEvent["event_type"].(string)) {
	case WatchSchemaChanges:
		form, err := a.getForm(ctx)
		if err != nil {
			l.Error("Failed to get form", zap.Error(err))
			// Don't abort, dispatch the event without this data.
		}
		formsEvent["form"] = form

	case WatchNewResponses:
		responses, err := a.listResponses(ctx)
		if err != nil {
			l.Error("Failed to list responses", zap.Error(err))
			// Don't abort, dispatch the event without this data.
		}
		formsEvent["response"] = lastResponse(responses)
	}

	// Convert the raw data to an AutoKitteh event.
	wrapped, err := sdktypes.WrapValue(formsEvent)
	if err != nil {
		l.Error("Failed to wrap Google Forms event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Google Forms event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: strings.ToLower(formsEvent["event_type"].(string)),
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
