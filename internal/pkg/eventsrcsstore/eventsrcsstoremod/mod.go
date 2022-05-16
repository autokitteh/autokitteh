package eventsrcsstoremod

import (
	"context"
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore/eventsrcsstorefactory"
	"github.com/autokitteh/autokitteh/pkg/starlarkutils"
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

	store, err := eventsrcsstorefactory.OpenString(
		context.Background(),
		nil,
		arg,
	)

	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return starlarkstruct.FromStringDict(
		starlarkutils.Symbol("eventsrcs"),
		map[string]starlark.Value{
			"add":            starlark.NewBuiltin("eventsrcs.add", makeAdd(store)),
			"update":         starlark.NewBuiltin("eventsrcs.update", makeUpdate(store)),
			"get":            starlark.NewBuiltin("eventsrcs.get", makeGet(store)),
			"list":           starlark.NewBuiltin("eventsrcs.list", makeList(store)),
			"add_binding":    starlark.NewBuiltin("eventsrcs.add_binding", makeAddBinding(store)),
			"update_binding": starlark.NewBuiltin("eventsrcs.update_binding", makeUpdateBinding(store)),
			"get_bindings":   starlark.NewBuiltin("eventsrcs.get_bindings", makeGetBindings(store)),
		},
	), nil
}

func makeAdd(s eventsrcsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			idArg       starlark.String
			settingsArg *starlark.Dict
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"id", &idArg,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		var pbsettings apieventsrc.EventSourceSettingsPB

		if err := starlarkutils.FromStarlark(settingsArg, &pbsettings); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		settings, err := apieventsrc.EventSourceSettingsFromProto(&pbsettings)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		if err := s.Add(context.Background(), apieventsrc.EventSourceID(idArg), settings); err != nil {
			return nil, fmt.Errorf("add: %w", err)
		}

		return starlark.None, nil
	}
}

func makeUpdate(s eventsrcsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			idArg       starlark.String
			settingsArg starlark.Dict
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"idArg", &idArg,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		var pbsettings apieventsrc.EventSourceSettingsPB

		if err := starlarkutils.FromStarlark(&settingsArg, &pbsettings); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		settings, err := apieventsrc.EventSourceSettingsFromProto(&pbsettings)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		if err := s.Update(context.Background(), apieventsrc.EventSourceID(idArg), settings); err != nil {
			return nil, fmt.Errorf("update: %w", err)
		}

		return starlark.None, nil
	}
}

func makeGet(s eventsrcsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var idArg starlark.String

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"id", &idArg,
		); err != nil {
			return nil, err
		}

		src, err := s.Get(context.Background(), apieventsrc.EventSourceID(idArg))
		if err != nil {
			return nil, fmt.Errorf("get: %w", err)
		}

		return starlarkutils.ToStarlark(src.PB())
	}
}

func makeList(s eventsrcsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			anameArg starlark.String
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"aaccount_name?", &anameArg,
		); err != nil {
			return nil, err
		}

		var aname *apiaccount.AccountName
		if anameArg != "" {
			aname_ := apiaccount.AccountName(anameArg)
			aname = &aname_
		}

		names, err := s.List(context.Background(), aname)
		if err != nil {
			return nil, fmt.Errorf("get: %w", err)
		}

		return starlarkutils.ToStarlark(names)
	}
}

func makeAddBinding(s eventsrcsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			srcID, pid, assoc, cfg, name starlark.String
			settingsArg                  starlark.Dict
			approved                     starlark.Bool
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"src_id", &srcID,
			"project_id", &pid,
			"name", &name,
			"assoc?", &assoc,
			"approved?", &approved,
			"cfg?", &cfg,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		var pbsettings apieventsrc.EventSourceProjectBindingSettingsPB

		if err := starlarkutils.FromStarlark(&settingsArg, &pbsettings); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		settings, err := apieventsrc.EventSourceProjectBindingSettingsFromProto(&pbsettings)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		if err := s.AddProjectBinding(context.Background(), apieventsrc.EventSourceID(srcID), apiproject.ProjectID(pid), string(name), string(assoc), string(cfg), bool(approved), settings); err != nil {
			return nil, fmt.Errorf("add: %w", err)
		}

		return starlark.None, nil
	}
}

func makeUpdateBinding(s eventsrcsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			srcid, pid, name starlark.String
			settingsArg      starlark.Dict
			approved         starlark.Bool
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"srcid", &srcid,
			"project_id", &pid,
			"name", &name,
			"approved", &approved,
			"settings", &settingsArg,
		); err != nil {
			return nil, err
		}

		var pbsettings apieventsrc.EventSourceProjectBindingSettingsPB

		if err := starlarkutils.FromStarlark(&settingsArg, &pbsettings); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		settings, err := apieventsrc.EventSourceProjectBindingSettingsFromProto(&pbsettings)
		if err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}

		if err := s.UpdateProjectBinding(context.Background(), apieventsrc.EventSourceID(srcid), apiproject.ProjectID(pid), string(name), bool(approved), settings); err != nil {
			return nil, fmt.Errorf("update: %w", err)
		}

		return starlark.None, nil
	}
}

func makeGetBindings(s eventsrcsstore.Store) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			srcidArg, pidArg, nameArg, assocArg starlark.String
			approvedOnly                        starlark.Bool
		)

		if err := starlark.UnpackArgs(
			b.Name(), args, kwargs,
			"srcid?", &srcidArg,
			"project_id?", &pidArg,
			"name?", &nameArg,
			"approved_only", &approvedOnly,
			"assoc?", assocArg,
		); err != nil {
			return nil, err
		}

		var srcid *apieventsrc.EventSourceID
		if srcidArg != "" {
			x := apieventsrc.EventSourceID(srcidArg)
			srcid = &x
		}

		var pid *apiproject.ProjectID
		if pidArg != "" {
			pid_ := apiproject.ProjectID(pidArg)
			pid = &pid_
		}

		bs, err := s.GetProjectBindings(context.Background(), srcid, pid, string(nameArg), string(assocArg), bool(approvedOnly))
		if err != nil {
			return nil, fmt.Errorf("get: %w", err)
		}

		r := make([]starlark.Value, len(bs))
		for i, b := range bs {
			if r[i], err = starlarkutils.ToStarlark(b.PB()); err != nil {
				return nil, fmt.Errorf("binding %d: %w", i, err)
			}
		}

		return starlark.NewList(r), nil
	}
}
