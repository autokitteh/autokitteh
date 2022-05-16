package internalplugins

import (
	starlibtime "github.com/qri-io/starlib/time"
	"go.starlark.net/starlark"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/starlarkplugin"
)

var Time = starlarkplugin.Plugin(
	"TODO",
	func() starlark.StringDict {
		return starlibtime.Module.Members
	}(),
)
