package langstarlark

import (
	"context"

	"go.starlark.net/starlark"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
)

func setTLSContext(t *starlark.Thread, ctx context.Context) { t.SetLocal("context", ctx) }
func getTLSContext(t *starlark.Thread) context.Context      { return t.Local("context").(context.Context) }

func setTLSEnv(t *starlark.Thread, env *lang.RunEnv) { t.SetLocal("env", env) }
func getTLSEnv(t *starlark.Thread) *lang.RunEnv      { return t.Local("env").(*lang.RunEnv) }
