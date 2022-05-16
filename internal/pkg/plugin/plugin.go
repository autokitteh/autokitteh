package plugin

import (
	"context"
	"errors"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
)

var ErrPluginNotFound = errors.New("plugin not found")

type Plugin interface {
	Describe(ctx context.Context) (*apiplugin.PluginDesc, error)

	// Returns nil if not found.
	Get(ctx context.Context, name string) (*apivalues.Value, error)

	GetAll(ctx context.Context) (map[string]*apivalues.Value, error)

	Call(ctx context.Context, v *apivalues.Value, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error)
}
