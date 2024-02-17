package runtime

import (
	"fmt"

	"go.starlark.net/starlark"
)

var Version = fmt.Sprintf("%d.0", starlark.CompilerVersion)
