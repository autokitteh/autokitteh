package internalplugins

import (
	starlibhttp "github.com/qri-io/starlib/http"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/starlarkplugin"
)

var HTTP = starlarkplugin.Plugin(
	"TODO",
	func() starlark.StringDict {
		mod, err := starlibhttp.LoadModule()
		if err != nil {
			panic(err)
		}

		root := mod["http"].(*starlarkstruct.Struct)

		dict := make(starlark.StringDict)

		root.ToStringDict(dict)

		return dict
	}(),
)
