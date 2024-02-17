package sdkservices

import (
	"context"
)

// SecretsService is a simple API for managing autokitteh user secrets.
// So far, this is limited to connections (managed by integrations).
type Secrets interface {
	// Create generates a new token to represent a connection's specified
	// key-value data, and associates them bidirectionally. If the same
	// request is sent N times, this method returns N different tokens.
	Create(ctx context.Context, scope string, data map[string]string, key string) (token string, err error)
	// Get retrieves a connection's key-value data based on the given token.
	// If the token isn’t found then we return an error.
	Get(ctx context.Context, scope string, token string) (data map[string]string, err error)
	// List enumerates all the tokens (0 or more) that are associated with a given
	// connection identifier. This enables autokitteh to dispatch/fan-out asynchronous
	// events that it receives from integrations through all the relevant connections.
	List(ctx context.Context, scope string, key string) (tokens []string, err error)
}
