package akmod

import (
	"context"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/pluginimpl"
)

func getCreds(get func(context.Context, string, string) ([]byte, error)) pluginimpl.SimplePluginMethodFunc {
	return func(
		ctx context.Context,
		args []*apivalues.Value,
		kwargs map[string]*apivalues.Value,
	) (*apivalues.Value, error) {
		var kind, name string

		if err := pluginimpl.UnpackArgs(
			args,
			kwargs,
			"kind", &kind,
			"name", &name,
		); err != nil {
			return nil, err
		}

		v, err := get(ctx, kind, name)
		if err != nil {
			return nil, err
		}

		return apivalues.Bytes(v), nil
	}
}
