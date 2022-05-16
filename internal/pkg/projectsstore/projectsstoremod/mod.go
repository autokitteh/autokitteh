package projectsstoremod

import (
	"context"
	"errors"
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstorefactory"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/projectsstorefactory"
	"github.com/autokitteh/autokitteh/pkg/starlarkutils"
)

var Module = &starlarkstruct.Module{
	Name: "projects",
	Members: starlark.StringDict{
		"open": starlark.NewBuiltin("projects.open", open),
		"path": starlark.NewBuiltin("projects.path", pathFunc),
	},
}

func open(
	thread *starlark.Thread,
	b *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	// TODO: separate arg for accounts store.
	var arg string
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &arg); err != nil {
		return nil, err
	}

	astore, err := accountsstorefactory.OpenString(context.Background(), nil, arg)
	if err != nil {
		return nil, fmt.Errorf("accounts store: %w", err)
	}

	store, err := projectsstorefactory.OpenString(context.Background(), nil, arg, astore)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return starlarkstruct.FromStringDict(
		starlarkutils.Symbol("projects"),
		map[string]starlark.Value{
			"create":    starlark.NewBuiltin("projects.create", makeCreate(store)),
			"update":    starlark.NewBuiltin("projects.update", makeUpdate(store)),
			"get":       starlark.NewBuiltin("projects.get", makeGet(store)),
			"batch_get": starlark.NewBuiltin("projects.batch_get", makeBatchGet(store)),
			"setup":     starlark.NewBuiltin("projects.setup", makeSetup(store)),
			"teardown":  starlark.NewBuiltin("projects.teardown", makeTeardown(store)),
		},
	), nil
}

func pathFunc(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var arg starlark.String
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &arg); err != nil {
		return nil, err
	}

	p, err := apiprogram.ParsePathString(string(arg))
	if err != nil {
		return nil, err
	}

	return starlarkutils.ToStarlark(p.PB())
}

func makeCreate(s projectsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			settingsArg starlark.Value
			aname       starlark.String
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"account_name", &aname,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		settings, err := projectSettingsFromStruct(settingsArg)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		id, err := s.Create(context.Background(), apiaccount.AccountName(aname), projectsstore.AutoProjectID, settings)
		if err != nil {
			return nil, fmt.Errorf("create: %w", err)
		}

		return starlark.String(id.String()), nil
	}
}

func makeUpdate(s projectsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			settingsArg starlark.Value
			id          starlark.String
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"id", &id,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		settings, err := projectSettingsFromStruct(settingsArg)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		if err := s.Update(context.Background(), apiproject.ProjectID(id), settings); err != nil {
			return nil, fmt.Errorf("create: %w", err)
		}

		return starlark.None, nil
	}
}

func projectSettingsFromStruct(v starlark.Value) (*apiproject.ProjectSettings, error) {
	var pbsettings apiproject.ProjectSettingsPB

	if err := starlarkutils.FromStarlark(v, &pbsettings); err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}

	return apiproject.ProjectSettingsFromProto(&pbsettings)
}

func projectToStruct(a *apiproject.Project) (starlark.Value, error) {
	return starlarkutils.ToStarlark(a.PB())
}

func makeBatchGet(s projectsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var ids *starlark.List

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "ids", &ids); err != nil {
			return nil, err
		}

		apiids := make([]apiproject.ProjectID, ids.Len())
		for i := 0; i < ids.Len(); i++ {
			id, ok := ids.Index(i).(starlark.String)
			if !ok {
				return nil, fmt.Errorf("id #%d is not a string", i)
			}

			apiids[i] = apiproject.ProjectID(id)
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
				x, err := projectToStruct(v)
				if err != nil {
					return nil, fmt.Errorf("project: %w", err)
				}
				_ = d.SetKey(starlark.String(k), x)
			}
		}

		return d, nil
	}
}

func makeGet(s projectsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var id starlark.String

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "id", &id); err != nil {
			return nil, err
		}

		a, err := s.Get(context.Background(), apiproject.ProjectID(id))
		if err != nil {
			if errors.Is(err, projectsstore.ErrNotFound) {
				return starlark.None, nil
			}

			return nil, fmt.Errorf("get: %w", err)
		}

		return projectToStruct(a)
	}
}

func makeSetup(s projectsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

func makeTeardown(s projectsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
