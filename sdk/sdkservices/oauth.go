package sdkservices

import (
	"context"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// OAuth is a generic implementation of 3-legged OAuth v2 flows, reusable by
// OAuth-based integrations and autokitteh user authentication. It assumes
// that the autokitteh server has a public address for callbacks, which
// allows callers of this service not to care about this requirement.
type OAuth interface {
	Register(ctx context.Context, id string, cfg *oauth2.Config, opts map[string]string) error
	Get(ctx context.Context, id string) (*oauth2.Config, map[string]string, error)
	StartFlow(ctx context.Context, id string, cid sdktypes.ConnectionID) (string, error)
	Exchange(ctx context.Context, id, state, code string) (*oauth2.Token, error)
}
