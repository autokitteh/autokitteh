package dispatcher

import (
	"errors"

	httpint "go.autokitteh.dev/autokitteh/integrations/http"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var specialTriggerTypes = map[sdktypes.IntegrationID]func(sdktypes.Trigger, sdktypes.Event) (bool, map[string]sdktypes.Value, error){
	httpint.IntegrationID: processHTTPTrigger,
}

func processSpecialTrigger(t sdktypes.Trigger, iid sdktypes.IntegrationID, event sdktypes.Event) (bool, map[string]sdktypes.Value, error) {
	if f, ok := specialTriggerTypes[iid]; ok {
		return f(t, event)
	}

	return true, nil, nil
}

func processHTTPTrigger(trigger sdktypes.Trigger, event sdktypes.Event) (bool, map[string]sdktypes.Value, error) {
	// Get expected path pattern from the trigger.
	pathValue, ok := trigger.Data()["pattern"]
	if !ok {
		// No pattern means we don't need to match anything.
		return true, nil, nil
	}

	if !pathValue.IsString() {
		return false, nil, errors.New("path in trigger data is not a string")
	}

	triggerPath := pathValue.GetString().Value()

	// Get actual URL from the event.
	urlValue, ok := event.Data()["url"]
	if !ok {
		return false, nil, errors.New("missing url in event data")
	}

	var url struct{ Path string }

	if err := urlValue.UnwrapInto(&url); err != nil {
		return false, nil, err
	}

	params, err := httpint.MatchPattern(triggerPath, url.Path)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			err = nil
		}

		return false, nil, err
	}

	return true, map[string]sdktypes.Value{
		"params": sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(params, sdktypes.NewStringValue),
		),
	}, nil
}
