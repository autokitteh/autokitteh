package langtools

import (
	"context"
	"fmt"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
)

func GetModuleDependencies(ctx context.Context, cat lang.Catalog, mod *apiprogram.Module) ([]*apiprogram.Path, error) {
	l, err := cat.Acquire(mod.Lang(), "")
	if err != nil {
		return nil, fmt.Errorf("lang: %w", err)
	}

	return l.GetModuleDependencies(ctx, mod)
}
