package akmod

import (
	"context"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/pluginimpl"
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
