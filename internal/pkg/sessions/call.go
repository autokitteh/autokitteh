package sessions

import (
	"context"

	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/plugin"

	"github.com/autokitteh/L"
)

func (s *Sessions) call(
	ctx context.Context,
	callv *apivalues.Value,
	plug plugin.Plugin,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	l := s.L.With("call", callv)

	callvCall := apivalues.GetCallValue(callv)
	if callvCall == nil {
		return nil, L.Error(l, "call to non-call value")
	}

	l.Debug("invoking", "name", callvCall.Name, "args", args, "kwargs", kwargs)

	v, err := plug.Call(ctx, callv, args, kwargs)

	l.Debug("returned", "err", err, "v", v)

	if v != nil {
		if err := apivalues.Walk(v, func(curr, _ *apivalues.Value, _ apivalues.Role) error {
			if currcv, ok := curr.Get().(apivalues.CallValue); ok {
				if currcv.Issuer == "" {
					if err := apivalues.SetCallIssuer(curr, callvCall.Issuer); err != nil {
						return L.Error(l, "set call issuer error", "err", err)
					}
				} else if !callvCall.Flags["allow_passing_call_values"] && callvCall.Issuer != currcv.Issuer { // [# allow_passing_call_values #]
					// don't let the plugin fool the session into calling another plugin's call value.
					return L.Error(l, "invalid issuer returned by call", "returned", currcv.Issuer, "expected", callvCall.Issuer)
				}
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return v, err
}
