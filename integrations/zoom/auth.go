package zoom

import (
	"context"

	"golang.org/x/oauth2"
)

// serverToken retrieves a Server-to-Server (2-legged OAuth) token, using the connection's
// internal app details (based on: https://developers.zoom.us/docs/internal-apps/s2s-oauth/).
func serverToken(_ context.Context, _ *privateApp) (*oauth2.Token, error) {
	return nil, nil // TODO(INT-247): Implement.
}
