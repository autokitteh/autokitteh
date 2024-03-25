package dispatcher

import (
	"errors"
	"fmt"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"github.com/julienschmidt/httprouter"
)

var specialTriggerTypes = map[string]func(sdktypes.Trigger, sdktypes.Event) (bool, map[string]sdktypes.Value, error){
	"http_route": processHTTPTrigger,
}

func processSpecialTrigger(t sdktypes.Trigger, event sdktypes.Event) (bool, map[string]sdktypes.Value, error) {
	if f, ok := specialTriggerTypes[t.Type()]; ok {
		return f(t, event)
	}

	return true, nil, nil
}

func processHTTPTrigger(trigger sdktypes.Trigger, event sdktypes.Event) (bool, map[string]sdktypes.Value, error) {
	pathValue, ok := trigger.Data()["path"]
	if !ok {
		return false, nil, errors.New("missing path in trigger data")
	}

	if !pathValue.IsString() {
		return false, nil, errors.New("path in trigger data is not a string")
	}

	triggerPath := pathValue.GetString().Value()

	if len(triggerPath) == 0 || triggerPath[0] != '/' {
		// httprouter requires the path to start with a slash.
		triggerPath = "/" + triggerPath
	}

	if len(triggerPath) != 0 && triggerPath[len(triggerPath)-1] == '*' {
		// httprouter doesn't like '*' at the end of the path, must be qualified with a name.
		triggerPath += "rest"
	}

	urlValue, ok := event.Data()["url"]
	if !ok {
		return false, nil, errors.New("missing url in event data")
	}

	var url struct{ Path string }

	if err := urlValue.UnwrapInto(&url); err != nil {
		return false, nil, err
	}

	if len(url.Path) == 0 || url.Path[0] != '/' {
		// httprouter requires the path to start with a slash.
		url.Path = "/" + url.Path
	}

	// TODO: This is probably an overkill, but will do for now. Need to extract
	//       the routing logic from httppath and simplify it to deal with
	//       a single route.

	var r httprouter.Router

	err := func() (err error) {
		// r.Handle has a tendency to panic if the route is invalid.
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("could not register route: %v", r)
				return
			}
		}()

		// We don't care about the method here as long as it's the same in Lookup and Handle.
		// We just want the to be registered.
		r.Handle("GET", triggerPath, func(http.ResponseWriter, *http.Request, httprouter.Params) {})

		return
	}()
	if err != nil {
		return false, nil, err
	}

	h, params, _ := r.Lookup("GET", url.Path)
	if h == nil {
		return false, nil, errors.New("could not find route")
	}

	return true, map[string]sdktypes.Value{
		"params": sdktypes.NewDictValueFromStringMap(
			kittehs.ListToMap(params, func(p httprouter.Param) (string, sdktypes.Value) {
				return p.Key, sdktypes.NewStringValue(p.Value)
			}),
		),
	}, nil
}
