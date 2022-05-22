package accountsstoremod

import (
	"context"
	"errors"
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstorefactory"
	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
	"github.com/autokitteh/starlarkutils"
)

var Module = &starlarkstruct.Module{
	Name: "accounts",
	Members: starlark.StringDict{
		"open": starlark.NewBuiltin("accounts.open", open),
	},
}

func open(
	thread *starlark.Thread,
	b *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var arg string
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &arg); err != nil {
		return nil, err
	}

	store, err := accountsstorefactory.OpenString(
		context.Background(),
		nil,
		arg,
	)

	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return starlarkstruct.FromStringDict(
		starlarkutils.Symbol("accounts"),
		map[string]starlark.Value{
			"create":    starlark.NewBuiltin("accounts.create", makeCreate(store)),
			"update":    starlark.NewBuiltin("accounts.update", makeUpdate(store)),
			"get":       starlark.NewBuiltin("accounts.get", makeGet(store)),
			"batch_get": starlark.NewBuiltin("accounts.batch_get", makeBatchGet(store)),
			"setup":     starlark.NewBuiltin("accounts.setup", makeSetup(store)),
			"teardown":  starlark.NewBuiltin("accounts.teardown", makeTeardown(store)),
		},
	), nil
}

func makeCreate(s accountsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			settingsArg starlark.Value
			name        starlark.String
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"name", &name,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		var pbsettings apiaccount.AccountSettingsPB

		if err := starlarkutils.FromStarlark(settingsArg, &pbsettings); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		settings, err := apiaccount.AccountSettingsFromProto(&pbsettings)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		if err := s.Create(context.Background(), apiaccount.AccountName(string(name)), settings); err != nil {
			return nil, fmt.Errorf("create: %w", err)
		}

		return starlark.None, nil
	}
}

func makeUpdate(s accountsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			name        starlark.String
			settingsArg starlark.Value
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"name", &name,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		var pbsettings apiaccount.AccountSettingsPB

		if err := starlarkutils.FromStarlark(settingsArg, &pbsettings); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		settings, err := apiaccount.AccountSettingsFromProto(&pbsettings)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		if err := s.Update(context.Background(), apiaccount.AccountName(string(name)), settings); err != nil {
			return nil, fmt.Errorf("create: %w", err)
		}

		return starlark.None, nil
	}
}

func accountToStruct(a *apiaccount.Account) (starlark.Value, error) {
	return starlarkutils.ToStarlark(a.PB())
}

func makeBatchGet(s accountsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var names *starlark.List

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "names", &names); err != nil {
			return nil, err
		}

		apiids := make([]apiaccount.AccountName, names.Len())
		for i := 0; i < names.Len(); i++ {
			name, ok := names.Index(i).(starlark.String)
			if !ok {
				return nil, fmt.Errorf("id #%d is not a string", i)
			}

			apiids[i] = apiaccount.AccountName(name)
		}

		as, err := s.BatchGet(context.Background(), apiids)
		if err != nil {
			return nil, fmt.Errorf("batch_get: %w", err)
		}

		d := starlark.NewDict(len(as))
		for k, v := range as {
			if v == nil {
				_ = d.SetKey(starlark.String(k), starlark.None)
			} else {
				x, err := accountToStruct(v)
				if err != nil {
					return nil, fmt.Errorf("account: %w", err)
				}
				_ = d.SetKey(starlark.String(k), x)
			}
		}

		return d, nil
	}
}

func makeGet(s accountsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var name starlark.String

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &name); err != nil {
			return nil, err
		}

		a, err := s.Get(context.Background(), apiaccount.AccountName(string(name)))
		if err != nil {
			if errors.Is(err, accountsstore.ErrNotFound) {
				return starlark.None, nil
			}

			return nil, fmt.Errorf("get: %w", err)
		}

		return accountToStruct(a)
	}
}

func makeSetup(s accountsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
			return nil, err
		}

		if err := s.Setup(context.Background()); err != nil {
			return nil, fmt.Errorf("setup: %w", err)
		}

		return starlark.None, nil
	}
}

func makeTeardown(s accountsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
			return nil, err
		}

		if err := s.Teardown(context.Background()); err != nil {
			return nil, fmt.Errorf("teardown: %w", err)
		}

		return starlark.None, nil
	}
}
