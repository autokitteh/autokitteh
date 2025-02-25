package ak

import (
	"errors"
	"fmt"
	"time"

	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/tls"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/values"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func LoadModule() starlark.StringDict {
	return starlark.StringDict{
		"start":                starlark.NewBuiltin("start", start),
		"subscribe":            starlark.NewBuiltin("subscribe", subscribe),
		"unsubscribe":          starlark.NewBuiltin("unsubscribe", unsubscribe),
		"next_event":           starlark.NewBuiltin("next_event", nextEvent),
		"is_deployment_active": starlark.NewBuiltin("is_deployment_active", IsDeploymentActive),
	}
}

func start(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		loc          string
		inputs, memo *starlark.Dict
	)

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "loc", &loc, "inputs?", &inputs, "memo?", &memo); err != nil {
		return nil, err
	}

	sdkLoc, err := sdktypes.ParseCodeLocation(loc)
	if err != nil {
		return nil, fmt.Errorf("loc: %w", err)
	}

	vctx := values.FromTLS(th)

	sdkInputs := make(map[string]sdktypes.Value, inputs.Len())
	if inputs != nil {
		for _, item := range inputs.Items() {
			if sdkInputs[item[0].String()], err = vctx.FromStarlarkValue(item[1]); err != nil {
				return nil, fmt.Errorf("value for %v: %w", item[0].String(), err)
			}
		}
	}

	sdkMemo := make(map[string]string, memo.Len())
	if memo != nil {
		for _, item := range memo.Items() {
			k, ok := item[0].(starlark.String)
			if !ok {
				return nil, errors.New("memo: key must be a string")
			}
			v, ok := item[1].(starlark.String)
			if !ok {
				return nil, errors.New("memo: value must be a string")
			}

			sdkMemo[k.GoString()] = v.GoString()
		}
	}

	tls := tls.Get(th)
	sid, err := tls.Callbacks.Start(tls.GoCtx, tls.RunID, sdkLoc, sdkInputs, sdkMemo)
	if err != nil {
		return nil, err
	}

	return starlark.String(sid.String()), nil
}

func subscribe(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, filter string

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "name", &name, "filter?", &filter); err != nil {
		return nil, err
	}

	tls := tls.Get(th)
	sid, err := tls.Callbacks.Subscribe(tls.GoCtx, tls.RunID, name, filter)
	if err != nil {
		return nil, err
	}

	return starlark.String(sid), nil
}

func unsubscribe(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var id string

	if err := starlark.UnpackArgs(bi.Name(), args, kwargs, "signal_id", &id); err != nil {
		return nil, err
	}

	tls := tls.Get(th)
	err := tls.Callbacks.Unsubscribe(tls.GoCtx, tls.RunID, id)
	if err != nil {
		return nil, err
	}

	return starlark.None, nil
}

func nextEvent(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var timeout starlark.Value

	if err := starlark.UnpackArgs(bi.Name(), nil, kwargs, "timeout?", &timeout); err != nil {
		return nil, err
	}

	sids, err := kittehs.TransformError(args, func(v starlark.Value) (string, error) {
		s, ok := v.(starlark.String)
		if !ok {
			return "", errors.New("signal_ids: value must be a list of strings")
		}

		return s.GoString(), nil
	})
	if err != nil {
		return nil, err
	}

	var duration time.Duration
	if timeout != nil {
		errInvalid := errors.New("timeout: value must be a valid integer or float")

		switch t := timeout.(type) {
		case starlark.NoneType:
		case starlark.Int:
			ui64, ok := t.Int64()
			if !ok {
				return nil, errInvalid
			}
			duration = time.Duration(ui64) * time.Second
		case starlark.Float:
			duration = time.Duration(float64(time.Second) * float64(t))
		default:
			return nil, errInvalid
		}
	}

	tls := tls.Get(th)
	v, err := tls.Callbacks.NextEvent(tls.GoCtx, tls.RunID, sids, duration)
	if err != nil {
		return nil, err
	}

	return values.FromTLS(th).ToStarlarkValue(v)
}

func IsDeploymentActive(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(bi.Name(), args, kwargs); err != nil {
		return nil, err
	}

	tls := tls.Get(th)

	active, err := tls.Callbacks.IsDeploymentActive(tls.GoCtx)
	if err != nil {
		return nil, err
	}

	return starlark.Bool(active), nil
}
