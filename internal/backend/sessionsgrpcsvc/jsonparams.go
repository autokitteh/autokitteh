package sessionsgrpcsvc

import (
	"encoding/json"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func unpackJSONObject(in string, outmap map[string]*sdktypes.ValuePB) error {
	var inmap map[string]any

	if err := json.Unmarshal([]byte(in), &inmap); err != nil {
		return sdkerrors.NewInvalidArgumentError(`json_object_input: %w`, err)
	}

	for k, v := range inmap {
		u, err := sdktypes.WrapValue(v)
		if err != nil {
			return sdkerrors.NewInvalidArgumentError(`json_object_input["%s"]: %w`, k, err)
		}

		outmap[k] = u.ToProto()
	}

	return nil
}
