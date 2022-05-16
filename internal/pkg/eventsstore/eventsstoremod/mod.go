package eventsstoremod

import (
	"context"
	"errors"
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apievent"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsstore"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsstore/eventsstorefactory"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langstarlark"
	"gitlab.com/softkitteh/autokitteh/pkg/starlarkutils"
)

var Module = &starlarkstruct.Module{
	Name: "events",
	Members: starlark.StringDict{
		"open": starlark.NewBuiltin("events.open", open),
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

	store, err := eventsstorefactory.OpenString(
		context.Background(),
		nil,
		arg,
	)

	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return starlarkstruct.FromStringDict(
		starlarkutils.Symbol("events"),
		map[string]starlark.Value{
			"add":                     starlark.NewBuiltin("events.add", makeAdd(store)),
			"get":                     starlark.NewBuiltin("events.get", makeGet(store)),
			"list":                    starlark.NewBuiltin("events.list", makeList(store)),
			"get_event_state":         starlark.NewBuiltin("events.get_event_state", getEventState(store)),
			"get_event_project_state": starlark.NewBuiltin("events.get_event_project_state", getEventProjectState(store)),
		},
	), nil
}

func makeAdd(s eventsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			srcid, typ, originalID, associationToken starlark.String
			dataArg                                  starlark.Dict
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"src_id", &srcid,
			"type", &typ,
			"original_id?", &originalID,
			"association_token?", &associationToken,
			"data", &dataArg,
		); err != nil {
			return nil, err
		}

		strdict := make(map[string]starlark.Value, dataArg.Len())
		for _, k := range dataArg.Keys() {
			v, _, _ := dataArg.Get(k)

			strk, ok := k.(starlark.String)
			if !ok {
				return nil, fmt.Errorf("data is not a string dict")
			}

			strdict[string(strk)] = v
		}

		data, err := langstarlark.NewNaiveValues().FromStringDict(strdict, nil)
		if err != nil {
			return nil, fmt.Errorf("error translating data: %w", err)
		}

		id, err := s.Add(context.Background(), apieventsrc.EventSourceID(srcid), string(associationToken), string(typ), string(originalID), data, nil)
		if err != nil {
			return nil, fmt.Errorf("create: %w", err)
		}

		return starlark.String(id.String()), nil
	}
}

func makeGet(s eventsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var id starlark.String

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "id", &id); err != nil {
			return nil, err
		}

		e, err := s.Get(context.Background(), apievent.EventID(id))
		if err != nil {
			if errors.Is(err, eventsstore.ErrNotFound) {
				return starlark.None, nil
			}

			return nil, fmt.Errorf("get: %w", err)
		}

		data := starlark.NewDict(len(e.Data()))
		for k, v := range e.Data() {
			vv, err := langstarlark.NewNaiveValues().ToStarlarkValue(v)
			if err != nil {
				return nil, fmt.Errorf("invalid value: %v, %w", k, err)
			}

			_ = data.SetKey(starlark.String(k), vv)
		}

		return starlarkutils.ToStarlark(e.PB())
	}
}

func makeList(s eventsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			pidArg         starlark.String
			ofsArg, lenArg starlark.Int
		)

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "project_id?", &pidArg, "ofsArg?", &ofsArg, "lenArg", &lenArg); err != nil {
			return nil, err
		}

		var pid *apiproject.ProjectID
		if pidArg != "" {
			pid_ := apiproject.ProjectID(pidArg)
			pid = &pid_
		}

		ln, _ := lenArg.Uint64()
		ofs, _ := ofsArg.Uint64()

		rs, err := s.List(context.Background(), pid, uint32(ln), uint32(ofs))
		if err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}

		elems := make([]starlark.Value, len(rs))
		for i, r := range rs {
			// TODO: states.

			if elems[i], err = langstarlark.NewNaiveValues().ToStarlarkValue(r.Event.AsValue()); err != nil {
				return nil, fmt.Errorf("#%d: %w", i, err)
			}
		}

		return starlark.NewList(elems), nil
	}
}

func getEventState(s eventsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var id starlark.String

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "id", &id); err != nil {
			return nil, err
		}

		log, err := s.GetState(context.Background(), apievent.EventID(id))
		if err != nil {
			if errors.Is(err, eventsstore.ErrNotFound) {
				return starlark.None, nil
			}

			return nil, fmt.Errorf("get: %w", err)
		}

		slog := make([]starlark.Value, len(log))
		for i, curr := range log {
			if slog[i], err = starlarkutils.ToStarlark(curr.PB()); err != nil {
				return nil, fmt.Errorf("log %d: %w", i, err)
			}
		}

		return starlark.NewList(slog), nil
	}
}

func getEventProjectState(s eventsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var id, pid starlark.String

		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "id", &id, "project_id", &pid); err != nil {
			return nil, err
		}

		log, err := s.GetStateForProject(context.Background(), apievent.EventID(id), apiproject.ProjectID(pid))
		if err != nil {
			if errors.Is(err, eventsstore.ErrNotFound) {
				return starlark.None, nil
			}

			return nil, fmt.Errorf("get: %w", err)
		}

		slog := make([]starlark.Value, len(log))
		for i, curr := range log {
			if slog[i], err = starlarkutils.ToStarlark(curr.PB()); err != nil {
				return nil, fmt.Errorf("log %d: %w", i, err)
			}
		}

		return starlark.NewList(slog), nil
	}
}
