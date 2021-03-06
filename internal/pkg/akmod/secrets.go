package akmod

import (
	"context"

	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/pluginimpl"
)

func getSecret(get func(context.Context, string) (string, error)) pluginimpl.SimplePluginMethodFunc {
	return func(
		ctx context.Context,
		args []*apivalues.Value,
		kwargs map[string]*apivalues.Value,
	) (*apivalues.Value, error) {
		var name string

		if err := pluginimpl.UnpackArgs(
			args,
			kwargs,
			"name", &name,
		); err != nil {
			return nil, err
		}

		v, err := get(ctx, name)
		if err != nil {
			return nil, err
		}

		return apivalues.String(v), nil
	}
}
