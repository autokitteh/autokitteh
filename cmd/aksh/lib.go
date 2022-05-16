package main

import (
	"os"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/accountsstore/accountsstoremod"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsrcsstore/eventsrcsstoremod"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsstore/eventsstoremod"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/projectsstore/projectsstoremod"
)

func initModules() {
	starlark.Universe["struct"] = starlark.NewBuiltin("struct", starlarkstruct.Make)
	starlark.Universe["module"] = starlark.NewBuiltin("module", starlarkstruct.MakeModule)

	starlark.Universe["accounts"] = accountsstoremod.Module
	starlark.Universe["projects"] = projectsstoremod.Module
	starlark.Universe["events"] = eventsstoremod.Module
	starlark.Universe["eventsrcs"] = eventsrcsstoremod.Module

	starlark.Universe["getenv"] = starlark.NewBuiltin("getenv", getenv)
}

func getenv(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, def string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs,
		"name", &name,
		"default?", &def,
	); err != nil {
		return nil, err
	}

	v, ok := os.LookupEnv(name)
	if !ok {
		v = def
	}

	return starlark.String(v), nil
}
